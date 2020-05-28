package concurrent

import (
	"time"
	"log"
)

type StartGroutineFn func(done <-chan interface{}, pulseInterval time.Duration) (
	heartbeat <-chan interface{})

func NewSteward(timeout time.Duration, fn StartGroutineFn) StartGroutineFn {
	return func(done <-chan interface{}, pulseInterval time.Duration) <-chan interface{} {
		heartbeat := make(chan interface{})

		go func() {
			defer close(heartbeat)
			// 监听存在问题的goroutine
			var wardDone chan interface{}
			var wardHeartBeat <-chan interface{}

			// 监听goroutine
			startWard := func() {
				wardDone = make(chan interface{})
				wardHeartBeat = fn(or(done, wardDone), timeout/2)
			}

			startWard()
			pulse := time.Tick(pulseInterval)

		loop:
			for {
				timeoutSignal := time.After(timeout)
				for {
					select {
					case <-pulse: // 发送心跳给监听者
						select {
						case heartbeat <- struct{}{}:
						default:
						}

					case <-wardHeartBeat: // 接受来自被监听的心跳
						continue loop

					case <-timeoutSignal: // 监控者超时, 停止当前的监控者, 重新启动新的监控者
						log.Println("steward: ward unhealthy; restarting")
						close(wardDone)
						startWard()
						continue loop

					case <-done:
						return
					}

				}
			}
		}()

		return heartbeat
	}
}

func doworkFn(done <-chan interface{}, nums ...int) (StartGroutineFn, <-chan interface{}) {
	intchanstream := make(chan (<-chan interface{})) // 2. bridge模式
	intstream := bridge(done, intchanstream)         // 1.将监控关闭的内容放回返回值, 并返回所有监控器用来交流数据的通道

	return func(done <-chan interface{}, pulseInterval time.Duration) (<-chan interface{}) {
		heartbeat := make(chan interface{}) // 3. 建立闭包控制器的启动和关闭
		intstream := make(chan interface{}) // 4. 向各通道与监控器交互数据的实例

		go func() {
			defer close(intstream)

			select {
			case intchanstream <- intstream: // 5.向数据交互作用的通道传入数据
			case <-done:
				return
			}

			pulse := time.Tick(pulseInterval)
			for {
			loop:
				for _, val := range nums {
					if val < 0 {
						log.Printf("negative value: %v\n", val) // 6. 不正常工作状态
						return
					}

					for {
						select {
						case <-pulse: // 对外的心跳
							select {
							case heartbeat <- struct{}{}:
							default:
							}

						case intstream <- val: // 心跳的返回值
							continue loop

						case <-done:
							return

						}
					}
				}
			}

		}()

		return heartbeat

	}, intstream
}
