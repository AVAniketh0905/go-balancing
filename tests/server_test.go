package tests

import (
	"context"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AVAniketh0905/go-balancing/intenral/server"
	"github.com/stretchr/testify/assert"
)

func echoTCPClient(t *testing.T, config server.Config) {
	conn, err := net.Dial("tcp", config.Addr())
	if err != nil {
		t.Errorf("Failed to connect to server at %s: %v", config.Addr(), err)
		return
	}
	defer conn.Close()
	t.Logf("Successfully connected to server at %s", config.Addr())
}

func TestMultipleServers(t *testing.T) {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	srv1 := &server.TCPServer{}
	srv2 := &server.TCPServer{}
	srv1.Init(3000, ctx, func(ctx context.Context, conn net.Conn) { defer conn.Close() })
	srv2.Init(3001, ctx, func(ctx context.Context, conn net.Conn) { defer conn.Close() })

	servers := []server.Server{srv1, srv2}

	wg.Add(len(servers))
	for _, s := range servers {
		go func(s server.Server) {
			defer wg.Done()
			if err := s.Run(); err != nil {
				t.Errorf("Server stopped: %v", err)
			}
		}(s)
	}

	time.Sleep(1 * time.Second)
	wg.Add(len(servers))

	for _, s := range servers {
		s := s.(*server.TCPServer)
		go func(s *server.TCPServer) {
			defer wg.Done()
			echoTCPClient(t, s.Config)
		}(s)
	}

	wg.Wait()

	if ctx.Err() != context.DeadlineExceeded {
		t.Fatalf("Expected context deadline exceeded, got: %v", ctx.Err())
	}

	t.Log("Test passed: servers shut down after context timeout")
}

// Weak test, numConns only works for server side not for client.
func TestNumConns(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := &server.TCPServer{}
	srv.Init(8080, ctx, func(ctx context.Context, conn net.Conn) {
		time.Sleep(100 * time.Millisecond)
	})

	go func() {
		if err := srv.Run(); err != nil {
			assert.NoError(t, err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// clients
	numClients := 5
	var wg sync.WaitGroup

	for i := 0; i < numClients; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", srv.Config.Addr())
			if err != nil {
				log.Printf("Client %d: Connection error: %v", id, err)
				return
			}

			// log.Printf("Client %d: Connected successfully", id)
			defer func() {
				conn.Close()
				// log.Printf("Client %d: Connection closed", id)
			}()
		}(i)
	}

	wg.Wait()

	time.Sleep(5 * time.Second) // wait for all clients to close

	if srv.NumConns() != 0 {
		t.Errorf("expected numConns to be 0, got %d", srv.NumConns())
	}
}

func TestTCPServerWithMaxConns(t *testing.T) {
	// Server setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := &server.TCPServer{}
	srv.Init(8080, ctx, func(ctx context.Context, conn net.Conn) {
		defer conn.Close()

		time.Sleep(500 * time.Millisecond)
	})

	// Start the server
	go func() {
		if err := srv.Run(); err != nil {
			assert.NoError(t, err)
		}
	}()

	// Allow server to start
	time.Sleep(1 * time.Second)

	// Simulate multiple clients
	numClients := 5
	var activeClients int32
	for i := 0; i < numClients; i++ {
		go func() {
			atomic.AddInt32(&activeClients, 1)
			conn, err := net.Dial("tcp", srv.Config.Addr())
			assert.NoError(t, err)
			defer conn.Close()

			// Keep the connection alive for a short time
			time.Sleep(2 * time.Second)
			atomic.AddInt32(&activeClients, -1)
		}()
	}

	// Wait for clients to finish
	time.Sleep(3 * time.Second)

	// Shutdown the server
	cancel()
	time.Sleep(1 * time.Second) // Allow time for cleanup

	// Check max connections in the map
	value, ok := server.MaxConnsDB.Load(srv.Config.Addr())
	if !ok {
		t.Fatalf("expected maxConnsMap to have entry for server %s", srv.Config.Addr())
	}

	maxConns := value.(int32)
	if maxConns != int32(numClients) {
		t.Fatalf("expected max connections to be %d, got %d", numClients, maxConns)
	}

	t.Logf("Max connections for server %s: %d", srv.Config.Addr(), maxConns)
}
