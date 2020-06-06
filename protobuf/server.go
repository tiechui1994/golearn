package main

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"golearn/protobuf/pb"
)

type ServiceServer struct {
	savedFutures []*pb.Feature
}

func (s *ServiceServer) GetFuture(ctx context.Context, pointer *pb.Point) (*pb.Feature, error) {
	for _, future := range s.savedFutures {
		if proto.Equal(future.Location, pointer) {
			return future, nil
		}
	}

	return &pb.Feature{Name: "", Location: pointer}, nil
}

func main() {
	for i := 1; i <= 7; i++ {
		fmt.Printf("index:%v   0x%02x \n", i, i<<3|2)
	}

	fmt.Printf("%#v\n", []byte("B"))
}
