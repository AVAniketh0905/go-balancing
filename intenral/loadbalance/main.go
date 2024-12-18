package loadbalance

import (
	"cmp"
	"fmt"
	"math/rand"
	"slices"
	"sync"

	"github.com/AVAniketh0905/go-balancing/intenral/utils"
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

type WeightedRoundRobin struct {
	mu sync.Mutex

	q utils.CyclicSlice[string]
}

func (wrr *WeightedRoundRobin) SelectBackend(servers []string, weights []int) (string, error) {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	if len(servers) == 0 {
		return "", fmt.Errorf("no backend servers available")
	}

	if wrr.q.IsNill() {
		type A struct {
			S string
			W int
		}
		combined := []A{}

		for i, s := range servers {
			a := A{S: s, W: weights[i]}
			combined = append(combined, a)
		}

		slices.SortFunc(combined, func(a, b A) int {
			return cmp.Compare(b.W, a.W) // ascending order
		})

		servers = []string{}
		for _, a := range combined {
			for i := 0; i < a.W; i++ {
				servers = append(servers, a.S)
			}
		}
		wrr.q.Init(servers)
	}

	return wrr.q.Get(), nil
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
