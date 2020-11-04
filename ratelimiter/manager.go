package ratelimiter

import (
	"log"
	"sync/atomic"
	"time"
)

const MaxUint = ^uint(0)

const MaxInt = int(MaxUint >> 1)

type Manager struct {
	errorChan    chan error
	releaseChan  chan *Token
	outChan      chan *Token
	inChan       chan struct{}
	needToken    int64
	activeTokens map[string]*Token
	limit        int
	makeToken    tokenFactory
}

func (m *Manager) Acquire() (*Token, error) {
	go func() {
		m.inChan <- struct{}{}
	}()

	select {
	case t := <-m.outChan:
		return t, nil
	case err := <-m.errorChan:
		return nil, err
	}
}

func (m *Manager) Release(t *Token) {
	if t.IsExpired() {
		go func() {
			m.releaseChan <- t
		}()
	}
}

// NewManager creates a manager type
func NewManager(conf *Config) *Manager {
	m := &Manager{
		errorChan:    make(chan error),
		outChan:      make(chan *Token),
		inChan:       make(chan struct{}),
		activeTokens: make(map[string]*Token),
		releaseChan:  make(chan *Token),
		needToken:    0,
		limit:        conf.Limit,
		makeToken:    NewToken,
	}

	if m.limit <= 0 {
		m.limit = MaxInt
	}

	if conf.TokenResetsAfter > 0 {
		m.runResetTokenTask(conf.TokenResetsAfter)
	}

	return m
}

func (m *Manager) incNeedToken() {
	atomic.AddInt64(&m.needToken, 1)
}

func (m *Manager) decNeedToken() {
	atomic.AddInt64(&m.needToken, -1)
}

func (m *Manager) awaitingToken() bool {
	return atomic.LoadInt64(&m.needToken) > 0
}

func (m *Manager) tryGenerateToken() {
	// panic if token factory is not defined
	if m.makeToken == nil {
		panic("ErrTokenFactoryNotDefined")
	}

	// cannot continue if limit has been reached
	if m.isLimitExceeded() {
		m.incNeedToken()
		return
	}

	token := m.makeToken()

	// Add token to active map
	m.activeTokens[token.ID] = token

	// send token to outChan
	go func() {
		m.outChan <- token
	}()
}

func (m *Manager) isLimitExceeded() bool {
	if len(m.activeTokens) >= m.limit {
		return true
	}
	return false
}

func (m *Manager) releaseToken(token *Token) {
	if token == nil {
		log.Print("unable to relase nil token")
		return
	}

	if _, ok := m.activeTokens[token.ID]; !ok {
		log.Printf("unable to relase token %s - not in use", token)
		return
	}

	if !token.IsExpired() {
		log.Printf("unable to relase token %s - has not expired", token)
		return
	}

	// Delete from map
	delete(m.activeTokens, token.ID)

	// process anything waiting for a rate limit
	if m.awaitingToken() {
		m.decNeedToken()
		go m.tryGenerateToken()
	}
}

// loops over active tokens and releases any that are expired
func (m *Manager) releaseExpiredTokens() {
	for _, token := range m.activeTokens {
		if token.IsExpired() {
			go func(t *Token) {
				m.releaseChan <- t
			}(token)
		}
	}
}

func (m *Manager) runResetTokenTask(resetAfter time.Duration) {
	go func() {
		ticker := time.NewTicker(resetAfter)
		for range ticker.C {
			for _, token := range m.activeTokens {
				if token.NeedReset(resetAfter) {
					go func(t *Token) {
						m.releaseChan <- t
					}(token)
				}
			}
		}
	}()
}
