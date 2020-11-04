package ratelimiter

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func ExecuteCommand(r1, r2 RateLimiter, cmds []string) {
	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	doWork := func(id int, cmd string) {
		// Acquire a rate limit token
		token1, err := r1.Acquire()
		token2, err := r2.Acquire()
		if err != nil {
			panic(err)
		}

		// Simulate some work
		d, _ := time.ParseDuration(cmd)
		fmt.Printf("Worker %d Command: %s \n", id, cmd)
		time.Sleep(d)
		fmt.Printf("Worker %d Done\n", id)
		r1.Release(token1)
		r2.Release(token2)
		wg.Done()
	}

	for i, v := range cmds {
		wg.Add(1)
		go doWork(i, v)
	}

	wg.Wait()
}
