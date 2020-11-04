package main

import (
	"flag"
	"fmt"
	"github.com/666ghost/go-rate-limiter/ratelimiter"
	"strings"
	"time"
)

var n, x int
var s string

func init() {
	flag.IntVar(&n, "n", 0, "max running commands")
	flag.IntVar(&x, "x", 0, "max commands per minute")

}
func main() {
	flag.Parse()

	if n <= 0 || x <= 0 {
		panic("-n and -x flags must be set and greater than 0")
	}

	r1, err := ratelimiter.NewFixedWindowRateLimiter(&ratelimiter.Config{
		Limit:         x,
		FixedInterval: 1 * time.Minute,
	})

	if err != nil {
		panic(err)
	}
	r2, err := ratelimiter.NewMaxConcurrencyRateLimiter(&ratelimiter.Config{
		Limit: n,
	})

	if err != nil {
		panic(err)
	}
	_, err = fmt.Scan(&s)
	if err != nil {
		panic(err)
	}
	cmds := strings.Split(s, ",")

	ratelimiter.ExecuteCommand(r1, r2, cmds)
	//var timeLimit int32 = 5
	/*
	 */
}
