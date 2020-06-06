package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"

	"golearn/protobuf/pb"
)

func main() {
	var varintMsg = &pb.VarintMsg{
		ArgI32:  0x41,
		ArgI64:  0x12345678,
		ArgUI32: 0x332211,
		ArgUI64: 0x998877,
		ArgSI32: -100,
		ArgSI64: -200,
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

	// repeated bool, int32, uint32, sint32  编码: 编号+类型 数组字节长度 值1,值2
	// 0a 02 01 00
	// 12 03 88 02 02
	// 1a 03 01 81 04
	// 22 03 8c 04 04

	// repeated string,bytes  编码: 编号+类型 字符串长度 字符串值, ...
	// 2a 02 41 41 2a 02 42 42 2a 03 41 42 43 2a 03 42 43 44
	// 32 05 48 65 6c 6c 6f 32 04 41 42 43 44

	// repeated Message  编码: 编号+类型 Message编码字节长度 Message编码, ...
	// 3a 07 08 81 02 10 02 18 01 3a 06 08 01 10 02 18 01
	var repeat = &pb.Repeat{
		ArgBoolList: []bool{true, false},
		ArgI32List:  []int32{0x0108, 0x02},
		ArgUI32List: []uint32{0x01, 0x0201},
		ArgSI32List: []int32{0x0106, 0x02},
		ArgStrList:  []string{"AA", "BB", "ABC", "BCD"},
		ArgByList:   [][]byte{[]byte("Hello"), []byte("ABCD")},
		ArgSimple: []*pb.Simple{
			{
				ArgI32:  0x00,
				ArgUI32: 0x00,
				ArgBool: true,
			},
			{
				ArgI32:  0x0101,
				ArgUI32: 0x0202,
				ArgBool: true,
			},
		},
	}

	data, _ = proto.Marshal(repeat)
	fd, _ = os.Create("./src/rpc/repeat.bin")
	fd.Write(data)
	fd.Sync()

	var mp = &pb.Map{
		ArgII: map[int32]int32{0x01: 0x012, 0x0201: 0x02},
		ArgUI: map[uint32]uint32{0x01: 0x01, 0x0201: 0x05},
		ArgSS: map[string]string{"AA": "BB", "ABC": "XY", "B": "HV", "WX": "ABC"},
		ArgSU: map[string]uint32{"A": 0x01, "B": 0x0102, "EC": 0x012},
	}

	data, _ = proto.Marshal(mp)
	fd, _ = os.Create("./src/rpc/map.bin")
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
