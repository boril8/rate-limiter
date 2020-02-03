package ratelimiter

import (
	"sync"
	"time"
)

type MultiLimiter struct {
	mu           *sync.RWMutex
	limitChanMap map[string]*singleLimiter
	ticker       *time.Ticker
	cancelChan   chan struct{}
}

type singleLimiter struct {
	chunk     int // size of increase
	limitChan chan struct{}
}

// NewMultiLimiter returns the MultiLimiter object similar to RateLimiter object. Calls and Burst control the limiter: calls is
// the number of calls to limit, and burst is the allowed burst (it fills up over the unused calls).
// The timeFrame allows to specify over what period the calls are spread and tolerance is a %
// of extra calls to be tolerated. Eg. tolerance of 5% would allow 105% calls per timeFrame.
// To use the MultiLimiter, you need to initialize Tenant. The allowed calls are filled up for all tenants peridically,
// each tenant can have a integer multiple of basic rate. Eg. if you have the basic rate of 100 calls per minute, then
// you can have tenant with 100, 200, or 300 etc. calls per minute allowed. Each can have a different buffer.
func NewMultiLimiter(calls, burst int, timeFrame time.Duration, tolerance float64) *MultiLimiter {
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
	limitChanMap := make(map[string]*singleLimiter)

	// prepare the object to be returned
	ml := &MultiLimiter{
		mu:           &sync.RWMutex{},
		limitChanMap: limitChanMap,
		ticker:       tick,
		cancelChan:   cChan,
	}

	// run the background fillup
	go func(mLimiter *MultiLimiter) {
		for {
			select {
			case <-tick.C:
				mLimiter.mu.RLock()
				for _, v := range mLimiter.limitChanMap {
					v.incTenant()
				}
				mLimiter.mu.RUnlock()
			case <-cChan:
				return
			default:
			}
		}
	}(ml)

	return ml
}

// Wait is the rate limiting call, use it to 'consume' limit
func (mrl *MultiLimiter) Wait(tenantID string) bool {
	mrl.mu.RLock()
	defer mrl.mu.RUnlock()

	singleLimiter, ok := mrl.limitChanMap[tenantID]
	if !ok {
		return false
	}

	<-singleLimiter.limitChan
	return true
}

// HasTenant checks if such tenant is set up
func (mrl *MultiLimiter) HasTenant(tenantID string) bool {
	_, ok := mrl.limitChanMap[tenantID]
	if !ok {
		return false
	}
	return true
}

// AddTenant adds a tenant with a config.
// It also updates a tenant limits with a new config if it exist.
func (mrl *MultiLimiter) AddTenant(tenantID string, burst, chunk int) {
	mrl.mu.Lock()
	defer mrl.mu.Unlock()

	slOld, ok := mrl.limitChanMap[tenantID]
	if ok {
		slOld.chunk = chunk
		newChan := make(chan struct{}, burst)
		close(slOld.limitChan)

		c := 0
		for v := range slOld.limitChan {
			c++
			if c > burst {
				break
			}
			newChan <- v
		}

		slOld.limitChan = newChan
		return
	}

	sl := singleLimiter{
		limitChan: make(chan struct{}, burst),
		chunk:     chunk,
	}

	mrl.limitChanMap[tenantID] = &sl
}

func (mrl *MultiLimiter) Stop() {
	close(mrl.cancelChan)
	mrl.ticker.Stop()
}

func (sl *singleLimiter) incTenant() {
	for i := 0; i < sl.chunk; i++ {
		select {
		case sl.limitChan <- struct{}{}:
		default:
		}
	}
}
