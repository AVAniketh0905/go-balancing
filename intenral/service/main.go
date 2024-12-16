package service

import (
	"fmt"

	"github.com/AVAniketh0905/go-balancing/intenral/server"
)

type ServiceType int

const (
	DataCollection ServiceType = iota
)

type Service struct {
	Type   ServiceType
	Server server.Server
}

func (s Service) Info() string {
	return s.Server.Info() + fmt.Sprintf("Type: %d", s.Type)
}

func (s Service) Run() error {
	return s.Server.Run()
}
