package concurrent

func Take(done, stream <-chan interface{}, num int) <-chan interface{} {
	results := make(chan interface{})
	go func() {
		defer close(results)

		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case results <- <-stream:
			}
		}
	}()
	return results
}

// 桥接的核心在于沟通的桥梁, 桥梁必须灵活可靠, 否则桥接会失败
func Bridge(done <-chan interface{}, chans <-chan <-chan interface{}) <-chan interface{} {
	stream := make(chan interface{})
	go func() {
		defer close(stream)
		for {
			var curchan <-chan interface{}
			select {
			case maybe, ok := <-chans:
				if !ok {
					return
				}
				curchan = maybe
			case <-done:
				return
			}

			for val := range OrDone(done, curchan) {
				select {
				case <-done:
				case stream <- val:
				}
			}
		}
	}()

	return stream
}
