package main

import (
	"fmt"
	"testing"
	"time"
)

func TestMock_After(t *testing.T) {
	m := NewMock()
	m.AfterFunc(5*time.Second, func() {
		fmt.Println("Hello")
	})
	m.Add(6 * time.Second)
}
