package ratelimiter

import (
	"time"
)

// RateLimiter defines two methods for acquiring and releasing tokens
type RateLimiter interface {
	Acquire() (*Token, error)
	Release(*Token)
}

type Config struct {
	Limit         int
	FixedInterval time.Duration

	Throttle time.Duration

	TokenResetsAfter time.Duration
}
