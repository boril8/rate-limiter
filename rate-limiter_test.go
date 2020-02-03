package ratelimiter

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 0, time.Second, 0)
	defer rl.Stop()

	t0 := time.Now()

	for n := 0; n < 20; n++ {
		rl.Wait()
	}

	el := time.Since(t0)

	if el > 2010*time.Millisecond {
		t.Errorf("Did not limit rate properly")
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	rl := NewRateLimiter(200, 0, time.Second, 0)
	defer rl.Stop()

	b.N = 100

	for n := 0; n < b.N; n++ {
		rl.Wait()
	}
}
