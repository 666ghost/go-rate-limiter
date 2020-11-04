package ratelimiter

import (
	"time"

	"github.com/segmentio/ksuid"
)

// token factory function creates a new token
type tokenFactory func() *Token

type Token struct {
	// The unique token ID
	ID string

	CreatedAt time.Time

	ExpiresAt time.Time
}

func NewToken() *Token {
	return &Token{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Time{},
	}
}

func (t *Token) IsExpired() bool {
	now := time.Now().UTC()
	return t.ExpiresAt.Before(now)
}

func (t *Token) NeedReset(resetAfter time.Duration) bool {
	if time.Since(t.CreatedAt) >= resetAfter {
		return true
	}
	return false
}
