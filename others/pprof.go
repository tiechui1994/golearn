package main

import (
	_ "net/http/pprof"
	"net/http"
	"log"
	"sync"
	"fmt"
	"time"
	"io/ioutil"
)

func main() {
	go func() {
		//ip:port 依据自己情况而定
		log.Println(http.ListenAndServe("localhost:8082", nil))
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go doTask(&wg)
	wg.Wait()
}

func doTask(wg *sync.WaitGroup) {
	for {
		for i := 0; i < 10000; i++ {
			Add("hello world ! just do it!")
			len := dohttpRequest()
			Add(fmt.Sprintf("response length:%v", len))
		}

		time.Sleep(3 * time.Second)
	}
	wg.Done()
}

var datas []string

func Add(str string) string {
	data := []byte(str)
	dataStr := string(data)
	datas = append(datas, dataStr)
	return dataStr
}

var result []string

func dohttpRequest() int {
	resp, err := http.Get("https://www.alibaba.com/")
	if err != nil {
		log.Printf("Get err:%v", err)
		return 0
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ReadAll err:%v", err)
		return 0
	}

	result = append(result, string(bytes))
	return len(bytes)
}
