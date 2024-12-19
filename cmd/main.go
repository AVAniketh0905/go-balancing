package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AVAniketh0905/go-balancing/intenral/instance"
	"github.com/AVAniketh0905/go-balancing/intenral/loadbalance"
	"github.com/AVAniketh0905/go-balancing/intenral/revproxy"
	"github.com/AVAniketh0905/go-balancing/intenral/server"
	"github.com/AVAniketh0905/go-balancing/intenral/service"
	"golang.org/x/sync/errgroup"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func InitiateServInstances() {
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

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configs := []int{10000, 10001, 10002}

	var servers []*server.TCPServer
	for _, port := range configs {
		tcpServer := &server.TCPServer{
			Config: server.NewConfig(port),
			Ctx:    ctx,
			HandlerFunc: func(ctx context.Context, conn net.Conn) {
				time.Sleep(200 * time.Millisecond)
			},
		}
		servers = append(servers, tcpServer)
	}

	serverAddrs := []string{}
	for _, srv := range servers {
		serverAddrs = append(serverAddrs, srv.Config.Addr())
	}

	lc := &loadbalance.LeastConnections{}
	lc.Init(serverAddrs)

	rp, err := revproxy.New(lc, serverAddrs)
	if err != nil {
		log.Fatalf("Error creating reverse proxy: %v", err)
	}

	rpConfig := server.NewConfig(9000)
	rpServer := &server.TCPServer{
		Config:      rpConfig,
		Ctx:         ctx,
		HandlerFunc: rp.HandlerFunc,
	}

	servers = append(servers, rpServer)

	var insts []*instance.Instance
	for _, srv := range servers {
		inst := instance.New(service.Service{
			Type:   service.DataCollection,
			Server: srv,
		})
		insts = append(insts, inst)
	}

	for _, inst := range insts {
		go func() {
			inst.MonitorState()
		}()
	}

	for _, inst := range insts {
		go func() {
			if err := inst.Start(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	time.Sleep(1 * time.Second)

	for _, srvAddr := range serverAddrs {
		server.MaxConnsDB.Store(srvAddr, int32(0))
	}

	time.Sleep(2 * time.Second)

	var wg sync.WaitGroup
	numClients := 200
	wg.Add(numClients)

	g := &errgroup.Group{}
	counter := int32(0)
	for i := 0; i < numClients; i++ {
		clientID := i
		g.Go(func() error {
			defer wg.Done()

			client, err := net.Dial("tcp", rpConfig.Addr())
			if err != nil {
				// log.Println(clientID, insts[len(insts)-1].State(), rpConfig.Addr())
				return fmt.Errorf("client %d dial error: %v", clientID, err)
			}
			defer client.Close()

			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

			atomic.AddInt32(&counter, 1)

			log.Printf("Client %d completed task.", clientID)
			return nil
		})

		// if i > 50 {
		// 	time.Sleep(100 * time.Millisecond)
		// }
	}

	time.Sleep(2 * time.Second)

	wg.Wait()

	log.Println("total clients served, ", atomic.LoadInt32(&counter))

	if err := g.Wait(); err != nil {
		log.Fatalf("Error in client simulation: %v", err)
	}

	for _, inst := range insts {
		go func() {
			if err := inst.Stop(); err != nil {
				log.Printf("Error stopping instance: %v", err)
			}
		}()
	}

	time.Sleep(10 * time.Second)
}
