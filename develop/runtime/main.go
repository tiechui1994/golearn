package main

import (
	"bytes"
	"fmt"
	"os"
)

var buf bytes.Buffer

func Write(s string) {
	buf.WriteString(s + "\n")
}

func main() {
	prefix := `package main`
	Write(prefix)
	Write("")
	start := "func sum("
	N := 3
	for i := 0; i < N; i++ {
		if i == N-1 {
			start += fmt.Sprintf("i%d int64", i)
		} else {
			start += fmt.Sprintf("i%d int64, ", i)
		}
	}
	start += ") int64 {"
	Write(start)
	Write("	return 0")
	Write("}")

	Write("")
	Write("func main() {")
	start = "	go sum("
	for i := 0; i < N; i++ {
		if i == N-1 {
			start += fmt.Sprintf("%d", i)
		} else {
			start += fmt.Sprintf("%d, ", i)
		}
	}
	start += ")"
	Write(start)
	Write("}")
	fmt.Println((N*8 + 7) &^ 7, 0x7d8)
	fmt.Println(buf.String())

	fd, _ := os.Create("./sum.go")
	fd.WriteString(buf.String())
}
