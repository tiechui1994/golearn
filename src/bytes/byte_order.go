package main

import (
	"unsafe"
	"fmt"
	"bytes"
	"encoding/binary"
)

type data struct {
	id      string
	header  map[string]string
	payload string
}

const Data_Len = int(unsafe.Sizeof(data{}))

func main() {
	var dataBytes = make([]byte, Data_Len)
	d := (*data)(unsafe.Pointer(&dataBytes[0]))
	d.id = "10"
	d.header = map[string]string{"A": "1"}
	d.payload = "payload"

	fmt.Printf("%v \n", dataBytes)

	var buf = new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, dataBytes)
	fmt.Printf("%v \n", buf.Bytes())

	buf.Reset()
	binary.Write(buf, binary.LittleEndian, dataBytes)
	fmt.Printf("%v \n", buf.Bytes())
}
