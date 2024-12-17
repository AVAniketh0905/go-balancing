package loadbalance

import (
	"fmt"
	"math/rand"
	"sync"
)

type Algorithm interface {
	SelectBackend(servers []string) (string, error)
}

type RoundRobin struct {
	mu      sync.Mutex
	current int
}

func (rr *RoundRobin) SelectBackend(servers []string) (string, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(servers) == 0 {
		return "", fmt.Errorf("no backend servers available")
	}

	// Select the next server in a circular fashion
	server := servers[rr.current]
	rr.current = (rr.current + 1) % len(servers)

	return server, nil
}

type Random struct{}

func (r *Random) SelectBackend(servers []string) (string, error) {
	if len(servers) == 0 {
		return "", fmt.Errorf("no backend servers available")
	}

	// Randomly select a server
	selected := servers[rand.Intn(len(servers))]

	return selected, nil
}
