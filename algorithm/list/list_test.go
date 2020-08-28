package list

import (
	"testing"
	"fmt"
	"time"
)

func TestReverseBetween(t *testing.T) {
	s := time.Now()
	head := &ListNode{
		Next: &ListNode{
			Next: &ListNode{
				Next: &ListNode{
					Next: &ListNode{
						Val: 5,
					},
					Val: 4,
				},
				Val: 3,
			},
			Val: 2,
		},
		Val: 1,
	}
	reverseBetween(head, 2, 4)
	fmt.Printf("%+v", time.Now().Sub(s).Nanoseconds())
}
