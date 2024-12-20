package revproxy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/AVAniketh0905/go-balancing/intenral/connection"
	"github.com/AVAniketh0905/go-balancing/intenral/loadbalance"
)

const REVPROXY_ADDR = "localhost:80"

type ReverseProxy struct {
	lb      loadbalance.Algorithm
	servers []string
}

func (rp *ReverseProxy) Handler(ctx context.Context, src connection.Conn) {
	server, err := rp.lb.SelectBackend()
	if err != nil {
		log.Printf("Error selecting backend server: %v", err)
		src.Close()
		return
	}

	dst, err := net.Dial("tcp", server)
	if err != nil {
		log.Fatalf("Error connecting to backend server %s: %v", server, err)
		src.Close()
		return
	}
	defer dst.Close()

	log.Printf("Proxy: connection established between %s and %s\n", src.RemoteAddr(), dst.RemoteAddr())

	once := &sync.Once{}
	go proxy(once, src, dst)
	proxy(once, dst, src)
}

func proxy(once *sync.Once, dst io.WriteCloser, src io.ReadCloser) {
	defer func() {
		once.Do(func() {
			log.Println("Proxy: closing connections")
			dst.Close()
			src.Close()
		})
	}()

	for r := bufio.NewReader(src); ; {
		msg, err := r.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Fatalf("Proxy read error: %v", err)
			}
		}

		if _, err := dst.Write([]byte(msg)); err != nil {
			log.Fatal(err)
		}
	}
}

func New(lb loadbalance.Algorithm, servers []string) (*ReverseProxy, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no backend servers provided")
	}

	return &ReverseProxy{
		lb:      lb,
		servers: servers,
	}, nil
}
