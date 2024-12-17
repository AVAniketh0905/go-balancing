package tests

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/AVAniketh0905/go-balancing/intenral/server"
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

	servers := []server.Server{
		server.TCPServer{
			Config: server.NewConfig(3000),
			Ctx:    ctx,
			HandlerFunc: func(ctx context.Context, conn net.Conn) {
				defer conn.Close()
			},
		},
		server.TCPServer{
			Config: server.NewConfig(3001),
			Ctx:    ctx,
			HandlerFunc: func(ctx context.Context, conn net.Conn) {
				defer conn.Close()
			},
		},
	}

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
		s := s.(server.TCPServer)
		go func(s server.TCPServer) {
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
