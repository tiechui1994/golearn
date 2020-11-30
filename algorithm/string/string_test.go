package string

import (
	"fmt"
	"testing"
)

func TestStrIP(t *testing.T) {
	ip := "2552551122"
	fmt.Println(stringIP(ip))
}

func TestFindSubstring(t *testing.T) {
	s := "wordgoodgoodgoodbestword"
	words := []string{"word", "good", "best", "word"}
	findSubstring(s, words)
}
