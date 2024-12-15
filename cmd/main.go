package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/AVAniketh0905/go-balancing/intenral/server"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	s := server.TCPServer{
		Config:  server.NewConfig(3000),
		Context: ctx,
		HandlerFunc: func(ctx context.Context, conn net.Conn) {
		},
	}
	defer cancel()

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
