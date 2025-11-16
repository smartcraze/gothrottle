package proxy

import (
	"sync/atomic"
)

// Balancer handles load balancing across multiple backend servers
// Currently implements round-robin, extensible for other strategies
type Balancer struct {
	counter uint64
	targets []string
}

// NewBalancer creates a new load balancer with the given targets
func NewBalancer(targets []string) *Balancer {
	return &Balancer{
		targets: targets,
		counter: 0,
	}
}

// Next returns the next target using round-robin strategy
func (b *Balancer) Next() string {
	if len(b.targets) == 0 {
		return ""
	}

	if len(b.targets) == 1 {
		return b.targets[0]
	}

	// Atomic increment for thread-safe round-robin
	n := atomic.AddUint64(&b.counter, 1)
	return b.targets[(n-1)%uint64(len(b.targets))]
}

// Targets returns all configured targets
func (b *Balancer) Targets() []string {
	return b.targets
}

// Count returns the number of targets
func (b *Balancer) Count() int {
	return len(b.targets)
}