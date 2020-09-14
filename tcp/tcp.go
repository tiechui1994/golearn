package main

import (
	"net"
	"bytes"
	"bufio"
	"strings"
)

func main() {
	net.DialTCP("", nil,nil)

	// ReadFrom
	_ = bytes.Buffer{}

	_ = bufio.Writer{}
	_ = bufio.ReadWriter{}

	_ = net.TCPConn{}

	// WriteTo
	_ = bytes.Reader{}
	_ = bytes.Buffer{}

	_ = bufio.Reader{}
	_ = bufio.ReadWriter{}

	_ = strings.Reader{}

	_ = net.Buffers{}

}