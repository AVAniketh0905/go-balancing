package connection

import (
	"fmt"
	"net"
)

type Conn interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
	RemoteAddr() string
}

type TCPConnection struct {
	Conn net.Conn
}

func (t *TCPConnection) Read(p []byte) (n int, err error) {
	return t.Conn.Read(p)
}

func (t *TCPConnection) Write(p []byte) (n int, err error) {
	return t.Conn.Write(p)
}

func (t *TCPConnection) Close() error {
	return t.Conn.Close()
}

func (t *TCPConnection) RemoteAddr() string {
	return t.Conn.RemoteAddr().String()
}

type UDPConnection struct {
	Conn *net.UDPConn
	Addr *net.UDPAddr // Peer address to respond to
}

func (u *UDPConnection) Read(p []byte) (n int, err error) {
	n, addr, err := u.Conn.ReadFromUDP(p)
	if err == nil {
		u.Addr = addr // Save the address for Write calls
	}
	return n, err
}

func (u *UDPConnection) Write(p []byte) (n int, err error) {
	if u.Addr == nil {
		return 0, fmt.Errorf("no remote address available to write")
	}
	return u.Conn.WriteToUDP(p, u.Addr)
}

func (u *UDPConnection) Close() error {
	return u.Conn.Close()
}

func (u *UDPConnection) RemoteAddr() string {
	if u.Addr != nil {
		return u.Addr.String()
	}
	return "unknown"
}
