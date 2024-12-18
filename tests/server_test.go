package tests

import (
	"context"
	"log"
	"net"
	"sync"
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
