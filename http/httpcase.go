package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
)

func main() {
	for i := 0; i < 6; i++ {
		var response *http.Response
		if i%2 == 0 {
			response, _ = http.Get("https://www.baidu.com")
		} else if i%2 == 1 {
			response, _ = http.Get("https://www.qq.com")
		}

		ioutil.ReadAll(response.Body)
	}

	fmt.Printf("G: %v\n", runtime.NumGoroutine())
}
