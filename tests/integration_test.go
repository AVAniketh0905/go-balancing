package tests

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AVAniketh0905/go-balancing/intenral/connection"
	"github.com/AVAniketh0905/go-balancing/intenral/instance"
	"github.com/AVAniketh0905/go-balancing/intenral/loadbalance"
	"github.com/AVAniketh0905/go-balancing/intenral/revproxy"
	"github.com/AVAniketh0905/go-balancing/intenral/server"
	"github.com/AVAniketh0905/go-balancing/intenral/service"
	"golang.org/x/sync/errgroup"
)

func TestIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tcp1 := &server.TCPServer{}
	tcp2 := &server.TCPServer{}

	tcp1.Init(8000, ctx, func(ctx context.Context, conn connection.Conn) {
		time.Sleep(100 * time.Millisecond)
	})
	tcp2.Init(8001, ctx, func(ctx context.Context, conn connection.Conn) {
		time.Sleep(200 * time.Millisecond)
	})

	servers := []*server.TCPServer{tcp1, tcp2}

	var insts []*instance.Instance
	for _, srv := range servers {
		inst := instance.NewTCP(srv.Config.Port(), ctx, service.DataCollection{})
		insts = append(insts, inst)
	}

	for _, inst := range insts {
		go func(i *instance.Instance) {
			i.MonitorState()
		}(inst)
	}

	g, _ := errgroup.WithContext(ctx)
	for _, inst := range insts {
		g.Go(func() error {
			return inst.Start()
		})
	}

	time.Sleep(10 * time.Second)

	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}

	g = &errgroup.Group{}
	// shuld fail as insts are already stopped
	for _, inst := range insts {
		g.Go(func() error {
			return inst.Stop()
		})
	}

	time.Sleep(5 * time.Second)

	if err := g.Wait(); err != nil && err.Error() != "already stopped" {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestIntegration_LoadBalancing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configs := []int{10000, 10001, 10002}

	var servers []*server.TCPServer
	for _, port := range configs {
		tcpServer := &server.TCPServer{
			Config: server.NewConfig(port),
			Ctx:    ctx,
			Handler: func(ctx context.Context, conn connection.Conn) {
				time.Sleep(50 * time.Millisecond)
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
		t.Fatalf("Error creating reverse proxy: %v", err)
	}

	rpConfig := server.NewConfig(9000)
	rpServer := &server.TCPServer{
		Config:  rpConfig,
		Ctx:     ctx,
		Handler: rp.Handler,
	}

	servers = append(servers, rpServer)

	var insts []*instance.Instance
	for _, srv := range servers {
		inst := instance.NewTCP(srv.Config.Port(), ctx, service.DataCollection{})
		insts = append(insts, inst)
	}

	var mainWG sync.WaitGroup

	for _, inst := range insts {
		go func() {
			inst.MonitorState()
		}()
	}

	for _, inst := range insts {
		mainWG.Add(1)
		go func() {
			defer mainWG.Done()
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
	numClients := 200 // reduce time in servers' handler func to accomodate more conns per socket

	g := &errgroup.Group{}
	counter := int32(0)
	for i := 0; i < numClients; i++ {
		clientID := i
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()

			client, err := net.Dial("tcp", rpConfig.Addr())
			if err != nil {
				// t.Println(clientID, insts[len(insts)-1].State(), rpConfig.Addr())
				return fmt.Errorf("client %d dial error: %v", clientID, err)
			}
			defer client.Close()

			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

			atomic.AddInt32(&counter, 1)

			t.Logf("Client %d completed task.", clientID)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Fatalf("Error in client simulation: %v", err)
	}

	wg.Wait()
	t.Log("total clients served, ", atomic.LoadInt32(&counter))

	// cancel()
	wg.Add(len(insts))

	g_ := &errgroup.Group{}
	for _, inst := range insts {
		g_.Go(func() error {
			defer wg.Done()
			return inst.Stop()
		})
	}

	wg.Wait()

	if err := g_.Wait(); err != nil {
		t.Fatal(err)
	}

	mainWG.Wait()

	t.Log("all goroutines completed")
}
