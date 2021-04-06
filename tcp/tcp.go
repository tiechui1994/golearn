package main

import (
	"context"
	"log"
	"net"
	"syscall"
	"time"
)

func main() {
	config := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				log.Println(syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR))
			})
			return err
		},
	}

	addr := &net.TCPAddr{Port: 1234, IP: net.ParseIP("0.0.0.0")}
	listen, err := config.Listen(context.Background(), "tcp4", addr.String())
	if err != nil {
		log.Println("Listen", err)
		return
	}

	listener := listen.(*net.TCPListener)
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Accept", err)
			continue
		}

		log.Println("remote addr", conn.RemoteAddr())

		go func(conn *net.TCPConn) {
			select {
			case <-time.After(15 * time.Second):
				conn.Close()
			}
		}(conn)
	}
}
