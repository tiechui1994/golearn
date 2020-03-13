package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/net/ipv4"
)

func main() {
	//prog := "mcast_listener"

	//if len(os.Args) != 5 {
	//	fmt.Printf("usage:   %s interface protocol group     address:port\n", prog)
	//	fmt.Printf("example: %s eth2      udp      224.0.0.9 0.0.0.0:2000\n", prog)
	//	return
	//}

	//ifname := os.Args[1]
	//proto := os.Args[2]
	//group := os.Args[3]
	//addrPort := os.Args[4]

	ifname := "enp0s31f6"
	proto := "udp"
	group := "224.0.0.9"
	addrPort := "0.0.0.0:2000"

	mcastRead(ifname, proto, group, addrPort)
}

func mcastRead(ifname, proto, group, addrPort string) {
	addr, port := splitHostPort(addrPort)
	p, err1 := strconv.Atoi(port)
	if err1 != nil {
		log.Fatal(err1)
	}

	a := net.ParseIP(addr)
	if a == nil {
		log.Fatal(fmt.Errorf("bad address: '%s'", addr))
	}

	g := net.ParseIP(group)
	if g == nil {
		log.Fatal(fmt.Errorf("bad group: '%s'", group))
	}

	ifi, err2 := net.InterfaceByName(ifname)
	if err2 != nil {
		log.Fatal(err2)
	}

	c, err3 := mcastOpen(a, p, ifname)
	if err3 != nil {
		log.Fatal(err3)
	}

	if err := c.JoinGroup(ifi, &net.UDPAddr{IP: g}); err != nil {
		log.Fatal(err)
	}

	if err := c.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true); err != nil {
		log.Fatal(err)
	}

	readLoop(c)

	c.Close()
}

func splitHostPort(hostPort string) (string, string) {
	s := strings.Split(hostPort, ":")
	host := s[0]
	if host == "" {
		host = "0.0.0.0"
	}
	if len(s) == 1 {
		return host, ""
	}
	return host, s[1]
}

func mcastOpen(bindAddr net.IP, port int, ifname string) (*ipv4.PacketConn, error) {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		log.Fatal(err)
	}
	if err := syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Fatal(err)
	}
	//syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
	if err := syscall.SetsockoptString(s, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, ifname); err != nil {
		log.Fatal(err)
	}

	lsa := syscall.SockaddrInet4{Port: port}
	copy(lsa.Addr[:], bindAddr.To4())

	if err := syscall.Bind(s, &lsa); err != nil {
		syscall.Close(s)
		log.Fatal(err)
	}
	f := os.NewFile(uintptr(s), "")
	c, err := net.FilePacketConn(f)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	p := ipv4.NewPacketConn(c)

	return p, nil
}

func readLoop(c *ipv4.PacketConn) {

	log.Printf("readLoop: reading")

	buf := make([]byte, 10000)

	for {
		n, cm, src, err1 := c.ReadFrom(buf)
		if err1 != nil {
			log.Printf("readLoop: ReadFrom: error %v", err1)
			break
		}

		var name string

		ifi, err2 := net.InterfaceByIndex(cm.IfIndex)
		if err2 != nil {
			log.Printf("readLoop: unable to solve ifIndex=%d: error: %v", cm.IfIndex, err2)
		}

		if ifi == nil {
			name = "ifname?"
		} else {
			name = ifi.Name
		}

		log.Printf("readLoop: recv %d bytes from %s to %s on %s", n, cm.Src, cm.Dst, name)

		c.WriteTo([]byte("Hello"),cm, src)
	}

	log.Printf("readLoop: exiting")
}
