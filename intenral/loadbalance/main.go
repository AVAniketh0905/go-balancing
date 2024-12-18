package loadbalance

import (
	"cmp"
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"sync"

	"github.com/AVAniketh0905/go-balancing/intenral/server"
	"github.com/AVAniketh0905/go-balancing/intenral/utils"
)

type Algorithm interface {
	SelectBackend() (string, error)
}

type RoundRobin struct {
	mu      sync.Mutex
	current int

	servers []string
}

func (rr *RoundRobin) Init(servers []string) {
	rr.servers = servers
}

func (rr *RoundRobin) SelectBackend() (string, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(rr.servers) == 0 {
		return "", fmt.Errorf("no backend rr.servers available")
	}

	// Select the next server in a circular fashion
	server := rr.servers[rr.current]
	rr.current = (rr.current + 1) % len(rr.servers)

	return server, nil
}

type WeightedRoundRobin struct {
	mu sync.Mutex
	q  utils.CyclicSlice[string]

	servers []string
	weights []int
}

func (wrr *WeightedRoundRobin) Init(servers []string, weights []int) {
	wrr.servers = servers
	wrr.weights = weights
}

func (wrr *WeightedRoundRobin) SelectBackend() (string, error) {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	if len(wrr.servers) == 0 {
		return "", fmt.Errorf("no backend servers available")
	}

	if wrr.q.IsNill() {
		type A struct {
			S string
			W int
		}
		combined := []A{}

		for i, s := range wrr.servers {
			a := A{S: s, W: wrr.weights[i]}
			combined = append(combined, a)
		}

		slices.SortFunc(combined, func(a, b A) int {
			return cmp.Compare(b.W, a.W) // ascending order
		})

		servers := []string{}
		for _, a := range combined {
			for i := 0; i < a.W; i++ {
				servers = append(servers, a.S)
			}
		}
		wrr.q.Init(servers)
	}

	return wrr.q.Get(), nil
}

type Random struct {
	servers []string
}

func (r *Random) Init(servers []string) {
	r.servers = servers
}

func (r *Random) SelectBackend() (string, error) {
	if len(r.servers) == 0 {
		return "", fmt.Errorf("no backend servers available")
	}

	// Randomly select a server
	selected := r.servers[rand.Intn(len(r.servers))]

	return selected, nil
}

// TODO
type LeastConnections struct {
	servers []string
}

func (lc *LeastConnections) Init(servers []string) {
	lc.servers = servers
}

func (lc *LeastConnections) SelectBackend() (string, error) {
	conns := []int32{}
	for _, s := range lc.servers {
		conn, ok := server.MaxConnsDB.Load(s)
		// log.Println("from db, ", s, conn)
		if !ok {
			return "", fmt.Errorf("couldnt find %v in global DB", s)
		}

		c, ok := conn.(int32)
		if !ok {
			return "", fmt.Errorf("improper conn int type, expected int32, got %v", reflect.TypeOf(conn))
		}

		conns = append(conns, c)
	}

	type A struct {
		S string
		C int32
	}

	combined := []A{}
	for i, s := range lc.servers {
		a := A{S: s, C: conns[i]}
		combined = append(combined, a)
	}

	idx := randomMinIdx(conns)

	return combined[idx].S, nil
}

func randomMinIdx(slice []int32) int {
	if len(slice) == 0 {
		return 0
	}

	min := slice[0]
	for _, val := range slice[1:] {
		if val < min {
			min = val
		}
	}

	var minIndices []int
	for i, val := range slice {
		if val == min {
			minIndices = append(minIndices, i)
		}
	}

	randomIndex := minIndices[rand.Intn(len(minIndices))]
	return randomIndex
}
