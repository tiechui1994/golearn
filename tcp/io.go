package main

import (
	"log"
	"io"
	"encoding/json"
	"net/http"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func post() {
	msg := struct {
		Name, Addr string
		Price      float64
	}{
		Name:  "hello",
		Addr:  "beijing",
		Price: 123.12,
	}

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		err := json.NewEncoder(w).Encode(msg)
		if err != nil {
			log.Println("encode", err)
		}
		log.Println("encode success")
	}()

	response, err := http.DefaultClient.Post("https://www.baidu.com", "application/json", r)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("hello")

	body := response.Body
	defer body.Close()
	log.Println("job success")
}


func main() {
	post()
}
