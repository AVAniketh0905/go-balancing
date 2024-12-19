package main

import (
	"context"
	"log"
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
	defer cancel()

	tcp1 := &server.TCPServer{}
	tcp2 := &server.TCPServer{}

	tcp1.Init(8000, ctx, func(ctx context.Context, conn net.Conn) {
		time.Sleep(100 * time.Millisecond)
	})
	tcp2.Init(8001, ctx, func(ctx context.Context, conn net.Conn) {
		time.Sleep(200 * time.Millisecond)
	})

	servers := []server.Server{tcp1, tcp2}

	var services []service.Service
	for _, s := range servers {
		srvc := service.Service{
			Type:   service.DataCollection,
			Server: s,
		}
		services = append(services, srvc)
	}

	var insts []*instance.Instance
	for _, srvc := range services {
		inst := instance.New(srvc)
		insts = append(insts, inst)
	}

	for _, inst := range insts {
		go func(i *instance.Instance) {
			i.MonitorState()
		}(inst)
	}

	for _, inst := range insts {
		go func(i *instance.Instance) {
			if err := i.Start(); err != nil {
				log.Fatal(err)
			}
		}(inst)
	}

	time.Sleep(1 * time.Second)

	for _, inst := range insts {
		go func(i *instance.Instance) {
			if err := i.Stop(); err != nil {
				log.Fatal(err)
			}
		}(inst)
	}

	time.Sleep(5 * time.Second)
}
