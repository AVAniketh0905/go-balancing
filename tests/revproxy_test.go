package tests

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/AVAniketh0905/go-balancing/intenral/revproxy"
	"github.com/AVAniketh0905/go-balancing/intenral/server"
	"github.com/stretchr/testify/assert"
)

func TestRevProxy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := server.NewConfig(3000)
	tcpServer := server.TCPServer{
		Config:  config,
		Context: ctx,
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

	rp, err := revproxy.New("tcp", config.Addr())
	assert.NoError(t, err)

	rpConfig := server.NewConfig(3470)
	rpServer := server.TCPServer{
		Config:      rpConfig,
		Context:     ctx,
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
