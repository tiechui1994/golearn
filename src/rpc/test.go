package main

import (
	"rpc/pb"
	"os"
	"github.com/golang/protobuf/proto"
	"fmt"
	"encoding/hex"
)

func main() {
	var varintMsg = &pb.VarintMsg{
		ArgI32:  0x41,
		ArgI64:  0x12345678,
		ArgUI32: 0x332211,
		ArgUI64: 0x998877,
		ArgSI32: -100,
		ArgSI64: -200,
		ArgBool: []bool{true, false},
		ArgEnum: pb.AuctionType_SECOND_PRICE,
	}
	data, _ := proto.Marshal(varintMsg)
	fd, _ := os.Create("./src/rpc/varint.bin")

	fmt.Printf("%v \n", hex.EncodeToString(data))
	fd.Write(data)
	fd.Sync()

	var bit64 = &pb.Bit64{
		ArgFixed64:  0x123456,
		ArgSFixed64: -100,
		ArgDouble:   3.1415926,
	}

	data, _ = proto.Marshal(bit64)
	fd, _ = os.Create("./src/rpc/bit64.bin")

	fmt.Printf("%v \n", hex.EncodeToString(data))
	fd.Write(data)
	fd.Sync()

	var bit32 = &pb.Bit32{
		ArgFixed32:  0x1234,
		ArgSFixed32: -10,
		ArgFloat:    3.1415,
	}

	var repeat = &pb.Repeat{
		ArgBoolList:  []bool{true, false},
		ArgI32List:   []int32{0x01, 0x02},
		ArgUI32List:  []uint32{0x01, 0x02},
		ArgSI32List:  []int32{0x01, 0x02},
		ArgStrList:   []string{"AA", "BB"},
		ArgVarintMsg: []*pb.VarintMsg{varintMsg},
	}

	data, _ = proto.Marshal(repeat)
	fd, _ = os.Create("./src/rpc/repeat.bin")
	fd.Write(data)
	fd.Sync()

	var payload = &pb.LenPayload{
		ArgMap:   map[string]int32{"A": 1, "B": 2, "C": 3,},
		ArgStr:   "Hello",
		ArgBytes: []byte("Hello"),
		ArgBit64: bit64,
		ArgBit32: bit32,
	}

	data, _ = proto.Marshal(payload)
	fd, _ = os.Create("./src/rpc/all.bin")
	fd.Write(data)
	fd.Sync()
}
