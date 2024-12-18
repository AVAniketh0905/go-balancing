package server

import (
	"context"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type MaxConnsMap struct {
	sync.Map
}

var MaxConnsDB *MaxConnsMap = &MaxConnsMap{}

func (m *MaxConnsMap) StoreMax(addr string, currentConns int32) {
	value, _ := m.LoadOrStore(addr, currentConns)
	if value.(int32) < currentConns {
		m.Store(addr, currentConns)
	}
}

type Server interface {
	NumConns() int32
	Info() string
	Run() error
}

type TCPServer struct {
	numConns int32

	Config      Config
	Ctx         context.Context
	HandlerFunc func(ctx context.Context, conn net.Conn)
}

func (s *TCPServer) Init(port int, ctx context.Context, handler func(ctx context.Context, conn net.Conn)) {
	s.Config = Config{port: port}
	s.Ctx = ctx
	s.HandlerFunc = handler
}

func (s *TCPServer) NumConns() int32 {
	return atomic.LoadInt32(&s.numConns)
}

func (s *TCPServer) Info() string {
	return "tcp:" + s.Config.Addr()
}

func (s *TCPServer) Run() error {
	addr := s.Config.Addr()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("l: ", err)
	}
	defer func() {
		l.Close()
		log.Println("server closed")
	}()
	log.Printf("server started at, %v\n...", addr)

	for {
		select {
		case <-s.Ctx.Done():
			log.Println("shutting down TCP server...")
			return nil
		default:
			l.(*net.TCPListener).SetDeadline(time.Now().Add(10 * time.Second)) // waits for 1 sec
			conn, err := l.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					// Timeout errors are expected; continue listening
					continue
				}
				log.Printf("error accepting connection: %v", err)
				return err
			}

			atomic.AddInt32(&s.numConns, 1)
			currentConns := s.NumConns()
			log.Println("server: conn started at, ", conn.LocalAddr(), ", numConns: ", currentConns)
			MaxConnsDB.StoreMax(s.Config.Addr(), currentConns)

			// go func(conns int32) {
			// 	// simulate some db call
			// 	log.Println("db call started...")
			// 	time.Sleep(5 * time.Second)
			// 	log.Println("db call ended, val stored", conns)
			// }(s.NumConns())

			go func() {
				defer func() {
					conn.Close()
					atomic.AddInt32(&s.numConns, -1)
				}()
				s.HandlerFunc(s.Ctx, conn)
			}()
		}
	}
}

type UDPServer struct {
	Config      Config
	Ctx         context.Context
	HandlerFunc func(ctx context.Context, conn net.PacketConn)
}

func (s *UDPServer) Init(port int, ctx context.Context, handler func(ctx context.Context, conn net.PacketConn)) {
	s.Config = Config{port: port}
	s.Ctx = ctx
	s.HandlerFunc = handler
}

func (s *UDPServer) NumConns() int32 {
	return 0

}

func (s *UDPServer) Info() string {
	return "udp" + s.Config.Addr()
}

func (s *UDPServer) Run() error {
	addr := s.Config.Addr()

	select {
	case <-s.Ctx.Done():
		log.Println("shutting down TCP server...")
		return nil
	default:
		conn, err := net.ListenPacket("udp", addr)
		if err != nil {
			log.Fatal("conn error", err)
			return err
		}
		log.Printf("listener started at, %v\n", addr)
		go func() {
			defer conn.Close()
			s.HandlerFunc(s.Ctx, conn)
		}()
	}

	return nil
}
