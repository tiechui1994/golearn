package concurrent

func Tee(done <-chan interface{}, in <-chan interface{}) (_, _ <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})

	go func() {
		defer close(out1)
		defer close(out2)
		for val := range OrDone(done, in) {
			var out1, out2 = out1, out2 // 1
			for i := 0; i < 2; i++ { // 2
				select {
				case <-done:
				case out1 <- val:
					out1 = nil //3
				case out2 <- val:
					out2 = nil //3
				}
			}
		}
	}()

	return out1, out2
}
