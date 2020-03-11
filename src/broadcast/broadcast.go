package main

import (
	"net"
	"log"
	"golang.org/x/net/ipv4"
)

// 文档: https://colobu.com/2016/10/19/Go-UDP-Programming/

// 多播
func Multicast() {
	//1. 获取一个interface
	en4, err := net.InterfaceByName("en4")
	if err != nil {
		log.Printf("InterfaceByName: %v", err)
		return
	}

	group := net.IPv4(224, 0, 0, 250)

	//2. bind本地地址
	c, err := net.ListenPacket("udp", "0.0.0.0:1024")
	if err != nil {
		log.Printf("ListenPacket:%v", err)
		return
	}
	defer c.Close()

	//3.
	p := ipv4.NewPacketConn(c)
	if err := p.JoinGroup(en4, &net.UDPAddr{IP: group}); err != nil {
		log.Printf("JoinGroup:%v", err)
		return
	}

	//4. 更多控制
	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
		log.Printf("SetControl: %v", err)
	}

	//5. 接收消息
	buf := make([]byte, 1500)
	for {
		n, cm, src, err := p.ReadFrom(buf)
		if err != nil {
			log.Printf("Read:%v", err)
			continue
		}

		if cm.Dst.IsMulticast() {
			if cm.Dst.Equal(group) {
				log.Printf("received: %s from <%s>", buf[:n], src.String())
				n, err = p.WriteTo([]byte("BroadCast"), cm, src)
				if err != nil {
					log.Printf("Write:%v", err)
					continue
				}
			} else {
				log.Printf("Unknown group")
				continue
			}
		}
	}
}

// 广播
func BroadCast() {
	listener, err := net.ListenUDP("udp",
		&net.UDPAddr{IP: net.IPv4zero, Port: 9981})
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Local: <%s> \n", listener.LocalAddr().String())
	data := make([]byte, 1024)
	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)
		if err != nil {
			log.Printf("error during read: %s", err)
		}
		log.Printf("<%s> %s\n", remoteAddr, data[:n])
		_, err = listener.WriteToUDP([]byte("world"), remoteAddr)
		if err != nil {
			log.Printf(err.Error())
		}
	}
}

func main() {
	BroadCast()
}
