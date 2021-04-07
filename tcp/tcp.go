package main

/*
#include <sys/socket.h>
*/
import "C"

import (
	"context"
	"log"
	"net"
	"syscall"
	"time"
)

const (
	SO_REUSEPORT = 0x0F
	SO_REUSEADDR = 0x02
)

func main() {
	config := net.ListenConfig{
		// 设置 TCP 选项(socket, tcp, ip)
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, SO_REUSEPORT, 1)
				err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_LINGER, 1)
				log.Println(syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR))
			})
			return err
		},
	}

	addr := &net.TCPAddr{Port: 1234, IP: net.ParseIP("127.0.0.1")}
	listen, err := config.Listen(context.Background(), "tcp4", addr.String())
	if err != nil {
		log.Println("Listen", err)
		return
	}

	log.Println(listen.Addr())

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
