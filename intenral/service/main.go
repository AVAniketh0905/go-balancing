package service

import (
	"context"

	"github.com/AVAniketh0905/go-balancing/intenral/connection"
)

type ServiceType interface {
	Name() string
	Handler(ctx context.Context, conn connection.Conn)
}

type DataCollection struct {
}

func (dc DataCollection) Name() string {
	return "data collection"
}

func (dc DataCollection) Handler(ctx context.Context, conn connection.Conn) {
	// log.Println("data collecting!")
}
