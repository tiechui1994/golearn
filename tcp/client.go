package main

import (
	"log"
	"net"
	"time"
)

func main() {
	addr := &net.TCPAddr{IP: net.ParseIP("192.168.1.118"), Port: 1234}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Println("Dail", err)
		return
	}

	done := time.After(10*time.Second)
	for {
		select {
		case <-done:
			conn.Close()
			log.Println("closing...")
		default:
			time.Sleep(5 * time.Second)
			log.Println(time.Now().Unix())
		}
	}
}
