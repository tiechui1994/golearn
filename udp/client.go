package main

import (
	"encoding/hex"
	"log"
	"net"
	"time"
)

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

// 广播的client
func BroadCastClient() {
	ip := net.ParseIP("255.255.255.255")
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 80}
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		log.Println(err)
	}

	buf := make([]byte, 10240)
	for i := 0; i < 10; i++ {
		origin := "5aa5aa555aa5aa55000000000000000000000000000000000000000000000000b1c20000000006000000000000000000"
		data, _ := hex.DecodeString(origin)
		n, err := conn.WriteToUDP(data, dstAddr)
		if err != nil {
			log.Println(err)
		}

		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Read [%s] from <%v>. Length:%v", hex.EncodeToString(buf[:n])[:10], addr.String(), n)
		time.Sleep(3 * time.Second)
	}
}

// 多播的client
func MulticastClient() {
	ip := net.ParseIP("192.168.50.14")
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 2000}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		log.Println("DialUDP:", err)
		return
	}
	defer conn.Close()

	var buf = make([]byte, 1024)
	_, err = conn.Write([]byte("hello"))
	if err != nil {
		log.Println("WriteToUDP:", err)
		return
	}

	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		log.Println("ReadFromUDP:", err)
		return
	}
	log.Printf("read from <%v> data [%v]", addr.String(), n)
}

func main() {
	MulticastClient()
}
