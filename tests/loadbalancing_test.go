package tests

import (
	"testing"

	"github.com/AVAniketh0905/go-balancing/intenral/loadbalance"
)

func TestWeightedRoundRobin(t *testing.T) {
	// Initialize the weighted round robin instance
	wrr := &loadbalance.WeightedRoundRobin{}

	// Test data: Servers and weights
	servers := []string{"A", "B", "C"}
	weights := []int{3, 1, 2} // A: 3, B: 1, C: 2

	expectedOrder := []string{"A", "A", "A", "C", "C", "B"} // Expected cyclic order

	// Initialize the backend selection
	for i := 0; i < len(expectedOrder); i++ {
		backend, err := wrr.SelectBackend(servers, weights)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if backend != expectedOrder[i] {
			t.Errorf("unexpected backend selected. Got %s, expected %s", backend, expectedOrder[i])
		}
	}

	// Test cyclic behavior
	for i := 0; i < len(expectedOrder); i++ {
		backend, err := wrr.SelectBackend(servers, weights)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify the cyclic continuation
		if backend != expectedOrder[i] {
			t.Errorf("unexpected cyclic backend selected. Got %s, expected %s", backend, expectedOrder[i])
		}
	}

	// Test empty servers
	_, err := wrr.SelectBackend([]string{}, []int{})
	if err == nil {
		t.Errorf("expected error when no servers are provided, but got nil")
	}
}
