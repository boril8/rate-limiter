package ratelimiter

import (
	"time"
)

type RateLimiter struct {
	limitChan  chan struct{}
	ticker     *time.Ticker
	cancelChan chan struct{}
}

// NewRateLimiter returns the RateLimiter object. Calls and Burst control the limiter: calls is
// the number of calls to limit, and burst is the allowed burst (it fills up over the unused calls).
// The timeFrame allows to specify over what period the calls are spread and tolerance is a %
// of extra calls to be tolerated. Eg. tolerance of 5% would allow 105% calls per timeFrame.
func NewRateLimiter(calls, burst int, timeFrame time.Duration, tolerance float64) *RateLimiter {
	if calls < 1 {
		panic("RateLimiter need positive number of calls")
	}

	if burst < 0 {
		burst = 0
	}

	// setup the configuration
	rate := time.Duration(float64(timeFrame) / (float64(calls) * (1 + tolerance)))
	tick := time.NewTicker(rate)
	cChan := make(chan struct{})
	lChan := make(chan struct{}, burst)

	// run the background fillup
	go func() {
		for {
			select {
			case <-tick.C:
				lChan <- struct{}{}
			case <-cChan:
				return
			default:
			}
		}
	}()

	// prepare the object to be returned
	return &RateLimiter{
		limitChan:  lChan,
		ticker:     tick,
		cancelChan: cChan,
	}
}

// Wait is the rate limiting call, use it to 'consume' limit
func (rl *RateLimiter) Wait() {
	<-rl.limitChan
}

func (rl *RateLimiter) Stop() {
	close(rl.cancelChan)
	rl.ticker.Stop()
}
