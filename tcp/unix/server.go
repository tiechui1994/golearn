package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func handle(conn net.Conn) {
	defer conn.Close()

	var (
		buffer bytes.Buffer
	)

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err == io.EOF {
			buffer.WriteString(msg)
			fmt.Println("send data", buffer.String())
			conn.Write(buffer.Bytes())
			buffer.Reset()
		} else if err != nil {
			fmt.Println("receive error: ", err)
			break
		}

		buffer.WriteString(msg)
	}
}

func main() {
	var file = "/tmp/unix.sock"
start:
	listen, err := net.Listen("unix", file)
	if err != nil {
		log.Println("UNIX Domain Socket:", err)
		os.Remove(file)
		goto start
	}

	fmt.Println("UNIX Domain Socket Success")
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("Accept", err)
			continue
		}

		go handle(conn)
	}
}
