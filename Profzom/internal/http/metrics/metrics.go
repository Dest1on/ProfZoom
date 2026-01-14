package metrics

import "sync/atomic"

type Collector struct {
	requests uint64
	errors   uint64
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) IncRequests() {
	atomic.AddUint64(&c.requests, 1)
}

func (c *Collector) IncErrors() {
	atomic.AddUint64(&c.errors, 1)
}

func (c *Collector) Snapshot() (uint64, uint64) {
	return atomic.LoadUint64(&c.requests), atomic.LoadUint64(&c.errors)
}
