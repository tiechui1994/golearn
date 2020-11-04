package main

import (
	"time"
	"strconv"
	"fmt"
	"math/rand"
)

// 通用的 senders 和 recvers 之间关闭通道的方法

// 代码逻辑:
// 终止条件(sender or recver) -> toStop -> close(stopCh) -> stop sender and recver
func _m_send_n_recv() {
	rand.Seed(time.Now().UnixNano())

	const Max = 2222
	const NumReceivers = 10
	const NumSenders = 1000

	// 数据流, sender -> recver
	dataCh := make(chan int, 100)

	// 停止信号的执行者
	stopCh := make(chan struct{})

	// 停止信号的发出者. It must be a buffered channel.
	toStop := make(chan string, 1)

	// moderator
	go func() {
		stoppedBy := <-toStop
		fmt.Println(stoppedBy)
		close(stopCh)
	}()

	// senders
	for i := 0; i < NumSenders; i++ {
		go func(id string) {
			for {
				value := rand.Intn(Max)

				// sender 的终止条件
				if value == 0 {
					select {
					case toStop <- "sender#" + id:
					default:
					}
					return
				}

				select {
				case <-stopCh:
					return
				case dataCh <- value:
				}
			}
		}(strconv.Itoa(i))
	}

	// receivers
	for i := 0; i < NumReceivers; i++ {
		go func(id string) {
			for {
				select {
				case <-stopCh:
					return
				case value := <-dataCh:

					// recver 的终止条件
					if value == Max-1 {
						select {
						case toStop <- "receiver#" + id:
						default:
						}
						return
					}
				}
			}
		}(strconv.Itoa(i))
	}

	time.Sleep(60 * time.Second)
}

// 代码逻辑:
// 终止条件(recver) -> stopCh -> stop sender
func _m_send_1_recv() {
	rand.Seed(time.Now().UnixNano())

	const Max = 100000
	const NumSenders = 1000

	// 数据流
	dataCh := make(chan int, 100)

	// 终止者
	stopCh := make(chan struct{})

	// senders
	for i := 0; i < NumSenders; i++ {
		go func() {
			for {
				select {
				case <-stopCh:
					return
				case dataCh <- rand.Intn(Max):
				}
			}
		}()
	}

	// the receiver
	go func() {
		for value := range dataCh {

			// 终止条件
			if value == Max-1 {
				fmt.Println("send stop signal to senders.")
				close(stopCh)
				return
			}

			fmt.Println(value)
		}
	}()

	time.Sleep(60 * time.Second)
}

// 代码逻辑:
// 终止条件(sender) -> stopCh -> stop recver
func _1_send_m_recv() {
	rand.Seed(time.Now().UnixNano())

	const Max = 100000
	const NumSenders = 1000

	// 数据流
	dataCh := make(chan int, 100)

	// 终止者
	stopCh := make(chan struct{})

	// senders
	go func() {
		for {
			val := rand.Intn(Max)
			if val == Max-1 {
				fmt.Println("send stop signal to recver.")
				close(stopCh)
				return
			}

			select {
			case dataCh <- rand.Intn(Max):
			}
		}
	}()

	// the receiver
	for i := 0; i < NumSenders; i++ {
		go func() {
			for {
				select {
				case value := <-dataCh:
					fmt.Println(value)
				case <-stopCh:
					return
				}
			}
		}()
	}

	time.Sleep(60 * time.Second)
}

func _copy() {
	type user struct {
		name string
		age  int8
	}

	var u = user{name: "Ankur", age: 25}
	var g = &u

	modifyUser := func(pu *user) {
		fmt.Println("modifyUser Received Vaule", pu)
		pu.name = "Anand"
	}

	printUser := func(u <-chan *user) {
		time.Sleep(2 * time.Second)
		fmt.Println("printUser goRoutine called", <-u)
	}

	c := make(chan *user, 5)
	c <- g
	fmt.Println(g)
	// modify g
	g = &user{name: "Ankur Anand", age: 100}
	go printUser(c)
	go modifyUser(g)
	time.Sleep(5 * time.Second)
	fmt.Println(g)
}

func main() {

}
