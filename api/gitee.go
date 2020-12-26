package main

import (
	"net/http"
	"net/url"
	"fmt"
	"io/ioutil"
	"regexp"
	"bytes"
	"encoding/json"
	"flag"
)

const (
	endpoint = "https://gitee.com"
	useraget = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
)

var (
	cookie    string
	project   string
	grouppath string
)

// 获取token
func getToken() (token string, err error) {
	u := endpoint
	request, _ := http.NewRequest("GET", u, nil)
	request.Header.Set("Cookie", cookie)
	request.Header.Set("User-Agent", useraget)

	response, err := http.Get(u)
	if err != nil {
		return token, err
	}

	data, _ := ioutil.ReadAll(response.Body)

	re := regexp.MustCompile(`<meta content="(.*?)" name="csrf-token"`)
	tokens := re.FindStringSubmatch(string(data))
	return tokens[0], nil
}

// groupath
func resources(token string) (err error) {
	u := endpoint + "/api/v3/internal/my_resources"
	request, _ := http.NewRequest("GET", u, nil)
	request.Header.Set("Cookie", cookie)
	request.Header.Set("X-CSRF-Token", token)
	request.Header.Set("User-Agent", useraget)

	response, err := http.DefaultClient.Do(request)
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

func forceSync(token string) (err error) {
	values := make(url.Values)
	values.Set("user_sync_code", "")
	values.Set("password_sync_code", "")
	values.Set("sync_wiki", "true")
	values.Set("authenticity_token", token)

	u := endpoint + "/" + grouppath + "/" + project + "/force_sync_project"
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(values.Encode()))
	request.Header.Set("Cookie", cookie)
	request.Header.Set("X-CSRF-Token", token)
	request.Header.Set("User-Agent", useraget)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode >= http.StatusOK {
		return nil
	}

	return fmt.Errorf("code:%v", response.StatusCode)
}

func main() {
	c := flag.String("cookie", "", "cookie value")
	p := flag.String("project", "", "project name")

	flag.Parse()

	if *c == "" {
		fmt.Println("不合法的cookie")
		return
	}
	if *p == "" {
		fmt.Println("不合法的project")
		return
	}

	cookie, project = *c, *p

	token, err := getToken()
	if err != nil {
		fmt.Println("网络存在问题")
		return
	}

	err = resources(token)
	if err != nil {
		fmt.Println("cookie内容不合法")
		return
	}

	err = forceSync(token)
	if err != nil {
		fmt.Printf("[%v] 同步失败\n", project)
		return
	}

	fmt.Printf("[%v] 同步成功", project)
}
