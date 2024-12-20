package instance

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/AVAniketh0905/go-balancing/intenral/server"
	"github.com/AVAniketh0905/go-balancing/intenral/service"
	"github.com/AVAniketh0905/go-balancing/intenral/utils"
)

type InstanceId int
type InstanceState int

const (
	Init InstanceState = iota
	Starting
	Running
	Stopped
)

type Instance struct {
	id  InstanceId
	srv server.Server

	port        int
	ctx         context.Context
	serviceType service.ServiceType

	mutex     sync.Mutex
	state     InstanceState
	stateChan chan InstanceState
}

func (i *Instance) Id() InstanceId {
	return i.id
}

func (i *Instance) Addr() string {
	config := server.NewConfig(i.port)
	return config.Addr()
}

func (i *Instance) State() InstanceState {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	return i.state
}

func (i *Instance) updateState(newState InstanceState) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.state = newState
	i.stateChan <- newState
}

func (i *Instance) Info() string {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	return fmt.Sprintf("Id: %d, State: %d, ", i.id, i.state) + i.srv.Info()
}

func (i *Instance) Start() error {
	i.updateState(Starting)

	err := i.srv.Run()
	if err != nil {
		i.updateState(Stopped)
	}
	return err
}

func (i *Instance) Run() error {
	i.updateState(Running)
	return i.srv.Run()
}

func (i *Instance) Stop() error {
	if i.State() == Stopped {
		return errors.New("already stopped")
	}
	i.updateState(Stopped)
	i.srv.Close()
	return nil
}

func (i *Instance) MonitorState() {
	if i.state == Stopped {
		return
	}

	for {
		select {
		case state := <-i.stateChan:
			fmt.Printf("Instance %d changed to state: %d\n", i.id, state)
		case <-i.ctx.Done():
			fmt.Printf("Instance %d context done, stopping state monitoring\n", i.id)
			return
		}
	}
}

func NewTCP(port int, ctx context.Context, serviceType service.ServiceType) *Instance {
	if port < 1025 || port > 65535 {
		log.Fatal("incorrect port rcvd, shuld lie btw 1025 & 65535", port)
	}

	srv := &server.TCPServer{}
	srv.Init(port, ctx, serviceType.Handler)

	return &Instance{
		id:          InstanceId(utils.GetRandNumber()),
		state:       Init,
		port:        port,
		ctx:         ctx,
		serviceType: serviceType,
		srv:         srv,
		stateChan:   make(chan InstanceState, 1), // Buffered channel to avoid blocking
	}
}

func NewUDP(port int, ctx context.Context, serviceType service.ServiceType) *Instance {
	if port < 1025 || port > 65535 {
		log.Fatal("incorrect port rcvd, shuld lie btw 1025 & 65535", port)
	}

	srv := &server.UDPServer{}
	srv.Init(port, ctx, serviceType.Handler)

	return &Instance{
		id:          InstanceId(utils.GetRandNumber()),
		state:       Init,
		port:        port,
		serviceType: serviceType,
		srv:         srv,
		stateChan:   make(chan InstanceState, 1), // Buffered channel to avoid blocking
	}
}
