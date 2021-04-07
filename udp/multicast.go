package main

import (
	"log"
	"net"
	"os"
	"syscall"

	"golang.org/x/net/ipv4"
)

var inetname string

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	inters, _ := net.Interfaces()
	for _, i := range inters {
		name := i.Name
		addrs, _ := i.Addrs()
		if name != "lo" && name != "docker0" && len(addrs) > 0 {
			inetname = name
			break
		}
	}
}

func main() {
	multicast()
}

func multicast() {
	// 1. interface
	inter, err := net.InterfaceByName(inetname)
	if err != nil {
		log.Fatal("InterfaceByName:", err)
	}

	// 2. open PacketConn
	pc, err := Open(net.IPv4zero, 2000, inetname)
	if err != nil {
		log.Fatal("mcastOpen:", err)
	}

	// 3. join
	group := net.IPv4(224, 0, 1, 10)
	if err := pc.JoinGroup(inter, &net.UDPAddr{IP: group}); err != nil {
		log.Fatal("JoinGroup:", err)
	}

	// 4. set controls
	if err := pc.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true); err != nil {
		log.Fatal("SetControlMessage:", err)
	}

	log.Println("reading......")
	defer pc.Close()

	// 5. read and write
	buf := make([]byte, 10240)
	for {
		n, cm, src, err1 := pc.ReadFrom(buf)
		if err1 != nil {
			log.Printf("ReadFrom: %v", err1)
			break
		}

		var name string

		ifi, err := net.InterfaceByIndex(cm.IfIndex)
		if err != nil {
			log.Printf("unable to solve ifIndex=%d: error: %v", cm.IfIndex, err)
		}

		if ifi == nil {
			name = "ifname?"
		} else {
			name = ifi.Name
		}

		log.Printf("recv %d bytes from %s to %s on %s", n, cm.Src, cm.Dst, name)

		pc.WriteTo([]byte("Hello"), cm, src)
	}

	log.Println("exiting....")

}

func Open(bindAddr net.IP, port int, ifname string) (*ipv4.PacketConn, error) {
	// socket, create what looks like an ordinary UDP socket
	socket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		log.Fatal("Socket:", err)
	}

	err = syscall.SetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		log.Fatal("SetsockoptInt:", err)
	}

	err = syscall.SetsockoptString(socket, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, ifname)
	if err != nil {
		log.Fatal("SetsockoptString:", err)
	}

	err = syscall.SetsockoptInt(socket, syscall.IPPROTO_IP, syscall.IP_MULTICAST_TTL, 64)
	if err != nil {
		log.Fatal("SetsockoptInt:", err)
	}

	// set up destination address
	addr := syscall.SockaddrInet4{Port: port}
	copy(addr.Addr[:], bindAddr.To4())

	// bind
	if err := syscall.Bind(socket, &addr); err != nil {
		syscall.Close(socket)
		log.Fatal("Bind:", err)
	}
	fd := os.NewFile(uintptr(socket), "")
	fpc, err := net.FilePacketConn(fd)
	fd.Close()
	if err != nil {
		log.Fatal("FilePacketConn:", err)
	}
	pc := ipv4.NewPacketConn(fpc)

	return pc, nil
}
