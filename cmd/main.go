package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/AVAniketh0905/go-balancing/intenral/instance"
	"github.com/AVAniketh0905/go-balancing/intenral/server"
	"github.com/AVAniketh0905/go-balancing/intenral/service"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	s := server.TCPServer{
		Config:  server.NewConfig(3000),
		Context: ctx,
		HandlerFunc: func(ctx context.Context, conn net.Conn) {
		},
	}
	defer cancel()

	myService := service.Service{
		Type:   service.DataCollection,
		Server: s,
	}
	myInstance := instance.New(myService)

	go myInstance.MonitorState()

	err := myInstance.Start()
	if err != nil {
		fmt.Printf("Failed to start instance: %v\n", err)
		return
	}

	time.Sleep(5 * time.Second)
	myInstance.Stop()

	time.Sleep(1 * time.Second)
}
