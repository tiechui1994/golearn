package main

func main() {
	f := func() func() uint64 {
		x := uint64(100)
		y := uint8(10)
		z := uint16(20)
		return func() uint64 {
			x += 100
			y += 1
			z += 1
			return x
		}
	}()

	f()
	f()
	f()
}
