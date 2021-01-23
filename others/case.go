package main

import (
	"regexp"
	"fmt"
)

func main() {
	r := regexp.MustCompile("\\d{3,}")
	str := "1234 1212 333"
	if r.MatchString(str) {
		indexs := r.FindAllStringIndex(str, 1)
		fmt.Println(indexs)
	}

}
