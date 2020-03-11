package main

import (
	"net"
	"log"
)

func Client() {
	ip := net.ParseIP("255.255.255.255")
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 9981}
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		log.Println(err)
	}
	n, err := conn.WriteToUDP([]byte("hello"), dstAddr)
	if err != nil {
		log.Println(err)
	}
	data := make([]byte, 1024)
	n, _, err = conn.ReadFrom(data)
	if err != nil {
		log.Println(err)
	}
	log.Printf("read [%s] from <%v>\n", data[:n], conn.RemoteAddr())
}

func main() {
	Client()
}
