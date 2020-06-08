package net

import (
	"fmt"
	"testing"
)

func TestListen(t *testing.T) {
	Listen()
}

type Addr interface {
	Addr() string
}

func IsNil(addr Addr) {
	if addr == nil {
		fmt.Println("is nil")
	}
}

func TestNil(t *testing.T) {
	IsNil(nil)
}
