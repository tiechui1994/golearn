package main

import (
	"context"
	"log"
	"net"
	"syscall"
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
				log.Println(syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR))
			})
			return err
		},
	}

	addr := &net.TCPAddr{Port: 5555, IP: net.IPv4zero}
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
			defer conn.Close()
			buf := make([]byte, 1024)
			for {
				n, err := conn.Read(buf)
				if err != nil {
					return
				}

				log.Println(string(buf[:n]))

				_, err = conn.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}(conn)
	}
}
