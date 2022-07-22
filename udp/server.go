package main

import (
	"fmt"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"sync"
)

func init() {
	log.SetPrefix("[server] ")
	log.SetFlags(log.Ltime | log.Lshortfile)
}

// 文档: https://colobu.com/2016/10/19/Go-UDP-Programming/

// 多播
// 首先找到要进行多播所使用的网卡, 然后监听本机合适的地址和服务端口.
// 将这个应用加入到多播组中, 它就可以从组中监听包信息, 当然你还可以对包传输进行更多的控制设置.
// 应用收到包后还可以检查包是否来自这个组的包.
func MultiCastCommon() {
	//1. 获取一个interface
	eth, err := net.InterfaceByName("enp0s31f6")
	if err != nil {
		log.Printf("InterfaceByName: %v", err)
		return
	}

	//2. bind本地地址
	c, err := net.ListenPacket("udp", "0.0.0.0:1024")
	if err != nil {
		log.Printf("ListenPacket:%v", err)
		return
	}
	defer c.Close()

	//3. 加入到组播地址当中.
	group := net.IPv4(224, 0, 0, 250)
	p := ipv4.NewPacketConn(c)
	if err := p.JoinGroup(eth, &net.UDPAddr{IP: group}); err != nil {
		log.Printf("JoinGroup:%v", err)
		return
	}

	//4. 更多控制
	if err := p.SetControlMessage(ipv4.FlagInterface, true); err != nil {
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
				log.Printf("received [%s] from <%s>", buf[:n], src.String())
				n, err = p.WriteTo([]byte("BroadCast"), cm, src)
				if err != nil {
					log.Printf("Write:%v", err)
					continue
				}
				log.Printf("send to <%s> success", src)
			} else {
				log.Printf("Unknown group")
				continue
			}
		}
	}
}

// 多播, 标准库提供的写法
// server 只能接收消息
func MultiCastServer() {
	addr, err := net.ResolveUDPAddr("udp", "224.0.0.250:2000")
	if err != nil {
		log.Println(err)
		return
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(interfaces))
	for i := range interfaces {
		ifi := interfaces[i]
		go func(ifi *net.Interface) {
			defer wg.Done()
			// 如果第二参数为nil, 它会使用系统指定多播接口,但是不推荐这样使用
			listener, err := net.ListenMulticastUDP("udp", ifi, addr)
			if err != nil {
				log.Println(err)
				return
			}

			log.Printf("interface: %q Local: <%s> \n", ifi.Name, listener.LocalAddr().String())

			data := make([]byte, 1024)
			for {
				n, remoteAddr, err := listener.ReadFromUDP(data)
				if err != nil {
					log.Printf("error during read: %s", err)
					continue
				}

				log.Printf("interface: %q, <%s> %s\n", ifi.Name, remoteAddr, data[:n])
			}
		}(&ifi)
	}
	wg.Wait()
}

// 广播 Server
func BroadCastServer() {
	listener, err := net.ListenUDP("udp",
		&net.UDPAddr{
			IP:   net.ParseIP("255.255.255.255"),
			Port: 80,
		},
	)
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
		_, err = listener.WriteToUDP(data[:n], remoteAddr)
		if err != nil {
			log.Printf(err.Error())
		}
	}
}

// 单播 Server
func SimpleCastServer() {
	listen, err := net.ListenUDP("udp4",
		&net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 8899,
		},
	)
	if err != nil {
		log.Println("ListenUDP", err)
		return
	}

	buf := make([]byte, 1024)
	for {
		// 注: ReadFromUDP 要记录 remoteAddr, 方便后面进行 WriteToUDP
		n, remoteAddr, err := listen.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("ReadFromUDP", err)
			continue
		}

		log.Printf("<%s> %s", remoteAddr, buf[:n])

		_, err = listen.WriteToUDP([]byte("ok"), remoteAddr)
		if err != nil {
			fmt.Println("WriteToUDP", err)
		}
	}
}

var (
	server  string
	servers = map[string]func(){
		"SimpleCastServer": SimpleCastServer,
		"MultiCastServer":  MultiCastServer,
		"BroadCastServer":  BroadCastServer,
	}
)

func main() {
	log.Printf("server: %v", server)
	if fun, ok := servers[server]; ok {
		fun()
	}
}
