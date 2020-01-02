package main

import (
	"context"
	"rpc/pb"
	"github.com/golang/protobuf/proto"
	"fmt"
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
	fmt.Printf("%x\n", 3<<3 | 1)
	fmt.Printf("%b\n", 100)
}
