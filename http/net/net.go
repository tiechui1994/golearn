package net

import (
	"fmt"
	"net"
)

func Listen() {
	listener, err := net.Listen("udp", "127.0.0.1:12345")
	fmt.Println(listener,err)
}
