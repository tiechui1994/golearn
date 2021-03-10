package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func handle(conn net.Conn) {
	defer conn.Close()

	size := 1024
	buf := make([]byte, size)
	buffer := bytes.Buffer{}
	for {
	read:
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("receive error: ", err)
			break
		}

		buffer.Write(buf[:n])
		if n == size {
			goto read
		}


		fmt.Println("receive data", buffer.String())
		if buffer.Len() == 0 {
			buffer.WriteString(time.Now().String())
		}
		conn.Write(buffer.Bytes())
		buffer.Reset()
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
