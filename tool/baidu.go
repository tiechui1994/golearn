package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

/**
百度语音转换测试API, 支持中文,英文
网站: https://developer.baidu.com/vcast
**/

func init() {
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 5 * time.Minute,
			}).DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

// https://developer.baidu.com/vcast 登录百度账号, 获取 Cookie 信息
var COOKIE = ""

func ConvertText(text string, filename string) (err error) {
	values := make(url.Values)
	values.Set("title", text)
	values.Set("content", ".")
	values.Set("sex", "0")    // 0 非情感女声, 1 非情感男声 3 情感男声  4 情感女声
	values.Set("speed", "6")  // 声速: 0-10
	values.Set("volumn", "8") // 声音大小
	values.Set("pit", "8")
	values.Set("method", "TRADIONAL")

	u := "https://developer.baidu.com/vcast/getVcastInfo"
	request, err := http.NewRequest("POST", u, bytes.NewBufferString(values.Encode()))
	if err != nil {
		log.Printf("NewRequest Failed: %v", err)
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	request.Header.Set("Cookie", COOKIE)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("Do Failed: %v", err)
		return err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("ReadAll Failed: %v", err)
		return err
	}

	log.Printf("uploadtext: %s", data)
	var res struct {
		BosUrl string `json:"bosUrl"`
		Status string `json:"status"`
	}

	json.Unmarshal(data, &res)
	if res.Status != "success" {
		log.Printf("failed: %v", string(data))
		return fmt.Errorf("")
	}

	return Download(res.BosUrl, filename)
}

func Download(u string, filename string) (err error) {
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Printf("NewRequest Failed: %v", err)
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("Do Failed: %v", err)
		return err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("ReadAll Failed: %v", err)
		return err
	}

	ioutil.WriteFile(filename+".mp3", data, 0666)

	return nil
}
