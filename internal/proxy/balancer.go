package proxy

import (
	"sync/atomic"
)

/*
Balancer handles load balancing across multiple backend servers using
round-robin strategy. Designed to be extensible for other algorithms.
*/
type Balancer struct {
	counter uint64
	targets []string
}

func NewBalancer(targets []string) *Balancer {
	return &Balancer{
		targets: targets,
		counter: 0,
	}
}

func (b *Balancer) Next() string {
	if len(b.targets) == 0 {
		return ""
	}

	if len(b.targets) == 1 {
		return b.targets[0]
	}

	n := atomic.AddUint64(&b.counter, 1)
	return b.targets[(n-1)%uint64(len(b.targets))]
}

func (b *Balancer) Targets() []string {
	return b.targets
}

func (b *Balancer) Count() int {
	return len(b.targets)
}