package main

import (
	"net/http"
	"net/url"
	"fmt"
	"bytes"
	"encoding/json"
	"flag"
	"time"
	"net"
	"net/http/cookiejar"
	"io/ioutil"
	"regexp"
)

const (
	endpoint = "https://gitee.com"
	useraget = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
)

var (
	cookie    string
	token     string
	project   string
	grouppath string
)

var gitee *http.Client

func init() {
	jar, _ := cookiejar.New(nil)
	gitee = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, time.Minute)
			},
		},
		Jar:     jar,
		Timeout: time.Minute,
	}
}

// CSRF-Token
func csrfToken() (err error) {
	u := endpoint
	request, _ := http.NewRequest("GET", u, nil)
	request.Header.Set("Cookie", cookie)
	request.Header.Set("User-Agent", useraget)

	response, err := gitee.Do(request)
	if err != nil {
		return err
	}

	data, _ := ioutil.ReadAll(response.Body)
	re := regexp.MustCompile(`meta content="(.*?)" name="csrf-token"`)
	tokens := re.FindStringSubmatch(string(data))
	if len(tokens) == 2 {
		token = tokens[1]
		return nil
	}

	return fmt.Errorf("invalid cookie")
}

// groupath
func resources() (err error) {
	u := endpoint + "/api/v3/internal/my_resources"
	request, _ := http.NewRequest("GET", u, nil)
	request.Header.Set("X-CSRF-Token", token)
	request.Header.Set("User-Agent", useraget)

	response, err := gitee.Do(request)
	if err != nil {
		return err
	}

	var result struct {
		GroupPath string `json:"groups_path"`
	}

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return err
	}

	grouppath = result.GroupPath
	return nil
}

// sync
func forceSync() (err error) {
	values := make(url.Values)
	values.Set("user_sync_code", "")
	values.Set("password_sync_code", "")
	values.Set("sync_wiki", "true")
	values.Set("authenticity_token", token)

	u := endpoint + "/" + grouppath + "/" + project + "/force_sync_project"
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(values.Encode()))
	request.Header.Set("X-CSRF-Token", token)
	request.Header.Set("User-Agent", useraget)

	response, err := gitee.Do(request)
	if err != nil {
		return err
	}

	data, _ := ioutil.ReadAll(response.Body)
	if len(data) == 0 {
		fmt.Printf("sync [%v] .... \n", project)
		time.Sleep(5 * time.Second)
		return forceSync()
	}

	var result struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}

	if result.Status == 1 {
		fmt.Printf("[%v] 同步成功\n", project)
	} else {
		fmt.Printf("[%v] 同步失败\n", project)
	}

	return nil
}

func main() {
	c := flag.String("cookie", "", "cookie value")
	p := flag.String("project", "", "project name")
	flag.Parse()

	if *c == "" {
		fmt.Println("未设置cookie")
		return
	}
	if *p == "" {
		fmt.Println("未设置的project")
		return
	}

	cookie, project = *c, *p

	err := csrfToken()
	if err != nil {
		fmt.Println("cookie内容不合法")
		return
	}

	err = resources()
	if err != nil {
		fmt.Println("cookie内容不合法")
		return
	}

	err = forceSync()
	if err != nil {
		fmt.Printf("[%v] 同步失败\n", project)
		return
	}
}
