package concurrent

func OrDone(done <-chan interface{}, stream <-chan interface{}) <-chan interface{} {
	vals := make(chan interface{})
	go func() {
		defer close(vals)

		for {
			select {
			case <-done:
				return
			case val, ok := <-stream:
				if ok == false {
					return
				}

				// 核心
				select {
				case <-done:
				case vals <- val:
				}

			}
		}

	}()
	return vals
}
