package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func post() {
	msg := struct {
		Cmd      string `json:"cmd"`
		Callback string `json:"callback"`
		Phone    string `json:"phone"`
	}{
		Cmd:      "1059",
		Callback: "phone",
		Phone:    "13152090953",
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

	request, _ := http.NewRequest("POST", "http://www.local.io/api", r)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	body := response.Body
	defer body.Close()

	data, _ := ioutil.ReadAll(body)
	log.Println("job success:", string(data))
}

func main() {
	post()
}
