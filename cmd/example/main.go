package main

import (
	"fmt"
	"time"

	ratelimiter "github.com/boril8/rate-limiter"
)

func main() {
	rl := ratelimiter.NewMultiLimiter(20, 0, time.Second, 0)
	defer rl.Stop()

	t0 := time.Now()

	rl.AddTenant("tenant1", 0, 1)
	rl.AddTenant("tenant2", 0, 1)

	for n := 0; n < 20; n++ {
		rl.Wait("tenant1")
	}

	for n := 0; n < 20; n++ {
		rl.Wait("tenant2")
	}

	el := time.Since(t0)

	fmt.Printf("Took %v\n", el)
}
