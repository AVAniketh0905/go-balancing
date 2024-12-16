package instance

import (
	"fmt"
	"sync"

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
	id      InstanceId
	state   InstanceState
	service service.Service

	stateChan chan InstanceState
	mutex     sync.Mutex
}

func (i *Instance) Id() InstanceId {
	return i.id
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
	return fmt.Sprintf("Id: %d, State: %d, ", i.id, i.state) + i.service.Info()
}

func (i *Instance) Start() error {
	i.updateState(Starting)

	go func() {
		err := i.Run()
		if err != nil {
			fmt.Printf("Instance %d encountered an error: %v\n", i.id, err)
			i.updateState(Stopped)
		}
	}()

	return nil
}

func (i *Instance) Run() error {
	i.updateState(Running)
	return i.service.Run()
}

func (i *Instance) Stop() error {
	i.updateState(Stopped)
	return nil
}

func (i *Instance) MonitorState() {
	for state := range i.stateChan {
		fmt.Printf("Instance %d changed to state: %d\n", i.id, state)
	}
}

func New(service service.Service) *Instance {
	return &Instance{
		id:        InstanceId(utils.GetRandNumber()),
		state:     Init,
		service:   service,
		stateChan: make(chan InstanceState, 1), // Buffered channel to avoid blocking
	}
}
