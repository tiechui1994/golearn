package main

import (
	"net"
	"log"
)

func Server() {
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: 80}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Printf("Listen:%v", err)
		return
	}

	buffer := make([]byte, 1024)
	for {
		n, client, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Read Faield:%v", err)
			continue
		}

		log.Printf("Receive %v Data:%v", client.String(), string(buffer[:n]))

		_, err = conn.WriteToUDP([]byte("Hello World"), client)
		if err != nil {
			log.Printf("Write Faield:%v", err)
		}
		log.Printf("Send To %v Success", client.String())
	}
}

func main() {
	Server()
}
