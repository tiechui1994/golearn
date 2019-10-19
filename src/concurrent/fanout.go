package concurrent

import "reflect"

// 将ch的内容完全复制N份给out
func FanOutAll(ch <-chan interface{}, out []chan interface{}) {
	go func() {
		defer func() {
			for i := range out {
				close(out[i])
			}
		}()

		for v := range ch {
			val := v
			for i := range out {
				out[i] <- val
			}
		}
	}()
}

// 构建发送类型的Cases, N个
// 每次都重新设置Chan和Send, Chan就是out的channel, Send是要发送的内容
// 循环便利所有的cases, 如果发送成功, 就将其对应的case的Chan标记为nil, 无法再发送内容,
// 直到当前值所有的Case都发送完毕
func FanOutAllReflect(ch <-chan interface{}, out []chan interface{}) {
	go func() {
		defer func() {
			for i := 0; i < len(out); i++ {
				close(out[i])
			}
		}()
		cases := make([]reflect.SelectCase, len(out))
		for i := range out {
			cases[i] = reflect.SelectCase{
				Dir: reflect.SelectSend, // 类型
			}
		}

		for v := range out {
			val := v
			for i := range cases {
				cases[i].Chan = reflect.ValueOf(out[i]) // chan, 用于接收或者发送的channel
				cases[i].Send = reflect.ValueOf(val)
			}

			for range cases { // for each channel
				chosen, _, _ := reflect.Select(cases)
				cases[chosen].Chan = reflect.ValueOf(nil)
			}
		}
	}()
}

// 将ch的内容发送到out, 所有out内容 = ch的内容

// 使用的是循环索引的方式进行发送
func FanOutSome(ch <-chan interface{}, out []chan interface{}) {
	go func() {
		defer func() {
			for i := range out {
				close(out[i])
			}
		}()

		var i int
		for v := range ch {
			val := v
			out[i] <- val
			i = (i + 1) % len(out)
		}
	}()
}

// 通过构建发送类型的Select的Case类型, 每次要发送的内容都设置到每个Case, 由Select()方法
// 随机挑选一个进行发送
func FanoutReflect(ch <-chan interface{}, out []chan interface{}) {
	go func() {
		defer func() {
			for i := 0; i < len(out); i++ {
				close(out[i])
			}
		}()
		cases := make([]reflect.SelectCase, len(out))
		for i := range out {
			cases[i] = reflect.SelectCase{
				Dir:  reflect.SelectSend,      // 类型
				Chan: reflect.ValueOf(out[i]), // chan, 用于接收或者发送的channel
			}
		}

		for v := range out {
			val := v
			// 将发送的值设置到每个case
			for i := range cases {
				cases[i].Send = reflect.ValueOf(val)
			}

			reflect.Select(cases)
		}
	}()
}
