package main

import (
	"log"
	"net"
	"os"
	"syscall"

	"golang.org/x/net/ipv4"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}
func main() {
	mcastRead()
}

func mcastRead() {
	group := net.IPv4(224, 0, 0, 9)

	ifi, err := net.InterfaceByName("wlp2s0")
	if err != nil {
		log.Fatal("InterfaceByName:", err)
	}

	pc, err := mcastOpen(net.IPv4zero, 2000, "wlp2s0")
	if err != nil {
		log.Fatal("mcastOpen:", err)
	}

	if err := pc.JoinGroup(ifi, &net.UDPAddr{IP: group}); err != nil {
		log.Fatal("JoinGroup:", err)
	}

	if err := pc.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true); err != nil {
		log.Fatal("SetControlMessage:", err)
	}

	readLoop(pc)

	pc.Close()
}

func mcastOpen(bindAddr net.IP, port int, ifname string) (*ipv4.PacketConn, error) {
	// socket, create what looks like an ordinary UDP socket
	socket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		log.Fatal("Socket:", err)
	}
	// allow multiple sockets to use the same PORT number
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

func readLoop(pc *ipv4.PacketConn) {

	log.Printf("readLoop: reading")

	buf := make([]byte, 10240)

	for {
		n, cm, src, err1 := pc.ReadFrom(buf)
		if err1 != nil {
			log.Printf("readLoop: ReadFrom: error %v", err1)
			break
		}

		var name string

		ifi, err := net.InterfaceByIndex(cm.IfIndex)
		if err != nil {
			log.Printf("readLoop: unable to solve ifIndex=%d: error: %v", cm.IfIndex, err)
		}

		if ifi == nil {
			name = "ifname?"
		} else {
			name = ifi.Name
		}

		log.Printf("readLoop: recv %d bytes from %s to %s on %s", n, cm.Src, cm.Dst, name)

		pc.WriteTo([]byte("Hello"), cm, src)
	}

	log.Printf("readLoop: exiting")
}
