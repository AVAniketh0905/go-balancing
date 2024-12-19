package tests

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/AVAniketh0905/go-balancing/intenral/instance"
	"github.com/AVAniketh0905/go-balancing/intenral/server"
	"github.com/AVAniketh0905/go-balancing/intenral/service"
	"golang.org/x/sync/errgroup"
)

func TestIntegration(t *testing.T) {
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

	g, _ := errgroup.WithContext(ctx)
	for _, inst := range insts {
		g.Go(func() error {
			return inst.Start()
		})
	}

	time.Sleep(10 * time.Second)

	if err := g.Wait(); err != nil && err != context.DeadlineExceeded {
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
