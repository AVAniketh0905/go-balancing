package server

import (
	"context"
	"log"
	"net"
	"time"
)

type Server interface {
	Info() string
	Run() error
}

type TCPServer struct {
	Config      Config
	Context     context.Context
	HandlerFunc func(ctx context.Context, conn net.Conn)
}

func (s TCPServer) Info() string {
	return "tcp:" + s.Config.Addr()
}

func (s TCPServer) Run() error {
	addr := s.Config.Addr()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("l: ", err)
	}
	defer l.Close()
	log.Printf("listener started at, %v\n", addr)

	for {
		select {
		case <-s.Context.Done():
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

			log.Println("conn started at, ", conn.LocalAddr())
			go func() {
				defer conn.Close()
				s.HandlerFunc(s.Context, conn)
			}()
		}
	}
}

type UDPServer struct {
	Config      Config
	Context     context.Context
	HandlerFunc func(ctx context.Context, conn net.PacketConn)
}

func (s UDPServer) Info() string {
	return "udp" + s.Config.Addr()
}

func (s UDPServer) Run() error {
	addr := s.Config.Addr()

	select {
	case <-s.Context.Done():
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
			s.HandlerFunc(s.Context, conn)
		}()
	}

	return nil
}
