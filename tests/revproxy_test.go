package tests

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/AVAniketh0905/go-balancing/intenral/loadbalance"
	"github.com/AVAniketh0905/go-balancing/intenral/revproxy"
	"github.com/AVAniketh0905/go-balancing/intenral/server"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestRevProxy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := server.NewConfig(3000)
	tcpServer := server.TCPServer{
		Config: config,
		Ctx:    ctx,
		HandlerFunc: func(ctx context.Context, conn net.Conn) {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					t.Fatalf("server read error: %v", err)
				}
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				t.Fatalf("server write error: %v", err)
			}
		},
	}

	go func() {
		err := tcpServer.Run()
		assert.NoError(t, err)
	}()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	servers := []string{tcpServer.Config.Addr()}

	randomLB := &loadbalance.Random{}
	randomLB.Init(servers)
	rp, err := revproxy.New(randomLB, servers)
	assert.NoError(t, err)

	rpConfig := server.NewConfig(3470)
	rpServer := server.TCPServer{
		Config:      rpConfig,
		Ctx:         ctx,
		HandlerFunc: rp.HandlerFunc,
	}

	go func() {
		err = rpServer.Run()
		assert.NoError(t, err)
	}()

	// Wait for rev proxy to start
	time.Sleep(1 * time.Second)

	client, err := net.Dial("tcp", rpConfig.Addr())
	assert.NoError(t, err)
	defer client.Close()

	// Send a message to the server via the proxy
	message := "hello, server!\n"
	_, err = client.Write([]byte(message))
	assert.NoError(t, err)

	// Read the response
	buf := make([]byte, 1024)
	n, err := client.Read(buf)
	assert.NoError(t, err)

	// Validate the response
	response := string(buf[:n])
	assert.Equal(t, message, response)
}

func TestLoadBalancing_RoundRobin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	config1 := server.NewConfig(8000)
	config2 := server.NewConfig(8001)
	config3 := server.NewConfig(8002)

	tcpServer1 := server.TCPServer{
		Config: config1,
		Ctx:    ctx,
		HandlerFunc: func(ctx context.Context, conn net.Conn) {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				t.Fatalf("server read error: %v", err)
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				t.Fatalf("server write error: %v", err)
			}
		},
	}

	tcpServer2 := server.TCPServer{
		Config: config2,
		Ctx:    ctx,
		HandlerFunc: func(ctx context.Context, conn net.Conn) {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				t.Fatalf("server read error: %v", err)
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				t.Fatalf("server write error: %v", err)
			}
		},
	}

	tcpServer3 := server.TCPServer{
		Config: config3,
		Ctx:    ctx,
		HandlerFunc: func(ctx context.Context, conn net.Conn) {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				t.Fatalf("server read error: %v", err)
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				t.Fatalf("server write error: %v", err)
			}
		},
	}

	go func() {
		err := tcpServer1.Run()
		assert.NoError(t, err)
	}()

	go func() {
		err := tcpServer2.Run()
		assert.NoError(t, err)
	}()

	go func() {
		err := tcpServer3.Run()
		assert.NoError(t, err)
	}()

	// Wait for servers to start
	time.Sleep(1 * time.Second)

	rr := &loadbalance.RoundRobin{}
	servers := []string{tcpServer1.Config.Addr(), tcpServer2.Config.Addr(), tcpServer3.Config.Addr()}
	rr.Init(servers)
	rp, err := revproxy.New(rr, servers)
	assert.NoError(t, err)

	rpConfig := server.NewConfig(9000)
	rpServer := server.TCPServer{
		Config:      rpConfig,
		Ctx:         ctx,
		HandlerFunc: rp.HandlerFunc,
	}

	go func() {
		err = rpServer.Run()
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	for i := 0; i < 5; i++ {
		client, err := net.Dial("tcp", rpConfig.Addr())
		assert.NoError(t, err)
		defer client.Close()

		// Send a message to the proxy
		message := fmt.Sprintf("hello, server %d!\n", i)
		_, err = client.Write([]byte(message))
		assert.NoError(t, err)

		// Read the response
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		assert.NoError(t, err)

		// Validate the response
		response := string(buf[:n])
		assert.Equal(t, message, response)
	}
}

func TestLoadBalancing_LeastConnections(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	config1 := server.NewConfig(10000)
	config2 := server.NewConfig(10001)
	config3 := server.NewConfig(10002)

	tcpServer1 := server.TCPServer{
		Config: config1,
		Ctx:    ctx,
		HandlerFunc: func(ctx context.Context, conn net.Conn) {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				t.Fatalf("server read error: %v", err)
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				t.Fatalf("server write error: %v", err)
			}
		},
	}

	tcpServer2 := server.TCPServer{
		Config: config2,
		Ctx:    ctx,
		HandlerFunc: func(ctx context.Context, conn net.Conn) {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				t.Fatalf("server read error: %v", err)
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				t.Fatalf("server write error: %v", err)
			}
		},
	}

	tcpServer3 := server.TCPServer{
		Config: config3,
		Ctx:    ctx,
		HandlerFunc: func(ctx context.Context, conn net.Conn) {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				t.Fatalf("server read error: %v", err)
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				t.Fatalf("server write error: %v", err)
			}
		},
	}

	go func() {
		err := tcpServer1.Run()
		assert.NoError(t, err)
	}()

	go func() {
		err := tcpServer2.Run()
		assert.NoError(t, err)
	}()

	go func() {
		err := tcpServer3.Run()
		assert.NoError(t, err)
	}()

	// Wait for servers to start
	time.Sleep(1 * time.Second)

	servers := []string{tcpServer1.Config.Addr(), tcpServer2.Config.Addr(), tcpServer3.Config.Addr()}

	for _, srv := range servers {
		server.MaxConnsDB.Store(srv, int32(0))
	}

	lc := &loadbalance.LeastConnections{}
	lc.Init(servers)

	rp, err := revproxy.New(lc, servers)
	assert.NoError(t, err)

	rpConfig := server.NewConfig(9000)
	rpServer := server.TCPServer{
		Config:      rpConfig,
		Ctx:         ctx,
		HandlerFunc: rp.HandlerFunc,
	}

	go func() {
		err = rpServer.Run()
		assert.NoError(t, err)
	}()

	time.Sleep(2 * time.Second)

	var wg sync.WaitGroup
	numClients := 50
	wg.Add(numClients)

	g := &errgroup.Group{}
	for i := 0; i < numClients; i++ {
		g.Go(func() error {
			defer wg.Done()

			client, err := net.Dial("tcp", rpConfig.Addr())
			assert.NoError(t, err)
			defer client.Close()

			// Send a message to the proxy
			message := fmt.Sprintf("client %d!\n", i)
			_, err = client.Write([]byte(message))
			if err != nil {
				return err
			}

			buf := make([]byte, 1024)
			_, err = client.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}

			// t.Logf("client %d - %v rcvd msg, %v", i, client.LocalAddr(), string(buf[:n]))

			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}

	time.Sleep(200 * time.Millisecond)

	wg.Wait()

	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}

	cancel()
}
