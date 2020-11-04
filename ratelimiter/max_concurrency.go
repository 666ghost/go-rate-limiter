package ratelimiter

import "errors"

func NewMaxConcurrencyRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.Limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}

	m := NewManager(conf)

	await := func() {
		go func() {
			for {
				select {
				case <-m.inChan:
					m.tryGenerateToken()
				case t := <-m.releaseChan:
					m.releaseToken(t)
				}
			}
		}()
	}

	await()
	return m, nil
}
