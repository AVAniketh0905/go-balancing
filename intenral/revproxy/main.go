package revproxy

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"sync"
)

const REVPROXY_ADDR = "localhost:80"

type ReverseProxy struct {
	dst        net.Conn
	remoteAddr string
}

func (rp ReverseProxy) HandlerFunc(ctx context.Context, src net.Conn) {
	log.Printf("Proxy: connection established between %s and %s\n", src.RemoteAddr(), rp.dst.RemoteAddr())
	defer func() {
		// log.Printf("Proxy: closing connection between %s and %s\n", src.RemoteAddr(), rp.dst.RemoteAddr())
		src.Close()
	}()

	once := &sync.Once{}
	go proxy(once, src, rp.dst)
	proxy(once, rp.dst, src)
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
			// log.Println("Proxy: connection closed by peer")
		}

		// log.Printf("Proxy forwarding: %s", msg)

		if _, err := dst.Write([]byte(msg)); err != nil {
			log.Fatal(err)
		}
	}
}

func New(network string, remoteAddr string) (*ReverseProxy, error) {
	dst, err := net.Dial(network, remoteAddr)
	if err != nil {
		return nil, err
	}

	return &ReverseProxy{
		dst:        dst,
		remoteAddr: remoteAddr,
	}, nil
}
