package concurrent

import "time"

func HeartInterval(done <-chan interface{}, pulseInterval time.Duration) (
	<-chan interface{}, <-chan time.Time) {
	heartbeat := make(chan interface{}) // 心跳信道
	results := make(chan time.Time)

	go func() {
		defer close(heartbeat)
		defer close(results)

		pulse := time.Tick(pulseInterval)       // 定时发送心跳, 每次心跳都意味着可以从该通道上读取到内容
		workGen := time.Tick(2 * pulseInterval) // 模拟进入的工作的另外一处代码

		sendPulse := func() {
			select {
			case heartbeat <- struct{}{}:
			default: // 必须考虑如果没有人接收到心跳的情况. 从goroutine发出的结果是至关重要的, 但心跳不是
			}
		}

		sendResult := func(r time.Time) {
			for {
				select {
				case <-done:
					return
				case <-pulse: // 像done一样, 无论何时执行发送或接收, 都需要考虑心跳发送的情况
					sendPulse()
				case results <- r:
					return
				}
			}
		}

		for {
			select {
			case <-done:
				return
			case <-pulse:
				sendPulse()
			case r := <-workGen:
				sendResult(r)
			}
		}
	}()
	return heartbeat, results
}

func HeartWork(done <-chan interface{}, num ...int) (
	<-chan interface{}, <-chan int) {
	heartbeat := make(chan interface{}, 1)
	results := make(chan int)
	go func() {
		defer close(heartbeat)
		defer close(results)

		time.Sleep(2 * time.Second)

		for _, i := range num {
			select {
			case heartbeat <- struct{}{}:
			default:
			}

			select {
			case <-done:
				return
			case results <- i:
			}
		}
	}()

	return heartbeat, results
}

func SecurtyHeartInterval(done <-chan interface{}, pulseInterval time.Duration, nums ...int) (
	<-chan interface{}, <-chan int) {
	heartbeat := make(chan interface{}, 1)
	results := make(chan int)

	go func() {
		defer close(heartbeat)
		defer close(results)

		time.Sleep(2 * time.Second) //模拟延迟(CPU,硬盘等)
		pulse := time.Tick(pulseInterval)

	loop:
		for _, n := range nums {
			for {
				select {
				case <-done:
					return
				case <-pulse:
					select {
					case heartbeat <- struct{}{}:
					default:
					}

				case results <- n:
					continue loop
				}
			}
		}
	}()
	return heartbeat, results
}
