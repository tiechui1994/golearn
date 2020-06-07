package net

import "net"

func Listen() {
	listener, _ := net.Listen("tcp", "127.0.0.1:12345")
	_ = listener
}
