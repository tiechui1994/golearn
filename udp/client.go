package main

import (
	"log"
	"net"
	"time"
)

func init() {
	log.SetPrefix("[client] ")
	log.SetFlags(log.Ltime | log.Lshortfile)
}

/**
// param network:
// The network must be a UDP network name; "udp", "udp4" (IPv4-only), "udp6" (IPv6-only)
//
// param raddr:
// If laddr is nil, a local address is automatically chosen.
// If the IP field of raddr is nil or an unspecified IP address, the local system is assumed.
//

//
// DialUDP acts like Dial for UDP networks.
//
DialUDP(network string, laddr, raddr *UDPAddr) (*UDPConn, error)


//
// ListenUDP acts like ListenPacket for UDP networks.
//
func ListenUDP(network string, laddr *UDPAddr) (*UDPConn, error)



//
// ListenMulticastUDP acts like ListenPacket for UDP networks but takes a group address on a specific network interface.
//
// ListenMulticastUDP listens on all available IP addresses of the local system including the group, multicast IP address.
//
// If ifi is nil, ListenMulticastUDP uses the system-assigned multicast interface, although this is not recommended because
// the assignment depends on platforms and sometimes it might require routing configuration.
//
// If the Port field of gaddr is 0, a port number is automatically chosen.
//
// There are golang.org/x/net/ipv4 and golang.org/x/net/ipv6 packages for general purpose uses.
//
func ListenMulticastUDP(network string, ifi *Interface, gaddr *UDPAddr) (*UDPConn, error)

*/

// 广播 UDP 通信. client 和 server 都需要 listen
// server 端需要 listen 255.255.255.255:80
func BroadCastClient() {
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP("255.255.255.255"), Port: 80}
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		log.Println(err)
	}

	data := []byte("Hello world")
	buf := make([]byte, 10240)
	for i := 0; i < 10; i++ {
		n, err := conn.WriteToUDP(data, dstAddr)
		if err != nil {
			log.Println(err)
		}

		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Read [%s] from <%v>. Length:%v", string(buf[:n]), addr.String(), n)
		time.Sleep(3 * time.Second)
	}
}

// 多播 UDP 通信. client 只能发送消息
// server 端需要 listen 0.0.0.0:2000, 然后将其加入都多播组 (例如, 224.0.0.250) 当中.
// server 端比较复杂
func MultiCastClient() {
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP("224.0.0.250"), Port: 2000}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		log.Println("DialUDP:", err)
		return
	}
	defer conn.Close()

	for {
		_, err = conn.Write([]byte(time.Now().String()))
		if err != nil {
			log.Println("WriteToUDP:", err)
			return
		}

		time.Sleep(time.Second)
	}
}

// 单播, 点对点的进行 UDP 通信
// server 端需要 listen 127.0.0.1:8899
func SimpleCastClient() {
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8899}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		log.Println(err)
	}

	data := []byte("Hello world")
	buf := make([]byte, 10240)
	for i := 0; i < 10; i++ {
		n, err := conn.Write(data)
		if err != nil {
			log.Println(err)
		}

		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Read [%s] from <%v>. Length:%v", string(buf[:n]), addr.String(), n)
		time.Sleep(3 * time.Second)
	}
}

var (
	client  string
	clients = map[string]func(){
		"SimpleCastClient": SimpleCastClient,
		"MultiCastClient":  MultiCastClient,
		"BroadCastClient":  BroadCastClient,
	}
)

func main() {
	log.Printf("client: %v", client)
	if fun, ok := clients[client]; ok {
		fun()
	}
}
