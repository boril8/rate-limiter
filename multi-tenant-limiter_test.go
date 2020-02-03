package ratelimiter

import (
	"testing"
	"time"
)

func TestMultiLimiter(t *testing.T) {
	rl := NewMultiLimiter(10, 0, time.Second, 0)
	defer rl.Stop()

	rl.AddTenant("tenant1", 0, 1)

	t0 := time.Now()

	for n := 0; n < 20; n++ {
		rl.Wait("tenant1")
	}

	el := time.Since(t0)

	if el > 2010*time.Millisecond {
		t.Errorf("Did not limit rate properly")
	}
}

func BenchmarkMultiLimiter(b *testing.B) {
	rl := NewMultiLimiter(200, 0, time.Second, 0)
	defer rl.Stop()

	b.N = 100

	for n := 0; n < b.N; n++ {
		rl.Wait("tenant1")
	}
}
