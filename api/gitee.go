package api

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"sync"
	"time"
)

/*
gitee 同步 git 项目
*/

const (
	endpoint = "https://gitee.com"
	useraget = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
)

var (
	sleep     time.Duration
	cookie    string
	token     string
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
	if grouppath == "" {
		return fmt.Errorf("invalid groupath")
	}

	return nil
}

// sync
func forceSync(project string) (err error) {
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
		time.Sleep(sleep * time.Second)
		return forceSync(project)
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

type sliceflag []string

func (s *sliceflag) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *sliceflag) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func main() {
	var p sliceflag
	flag.Var(&p, "project", "project name")
	c := flag.String("cookie", "", "cookie value")
	t := flag.Int("sleep", 3, "sync wait seconds")
	flag.Parse()

	if *c == "" {
		fmt.Println("未设置cookie")
		return
	}
	if len(p) == 0 {
		fmt.Println("未设置的project")
		return
	}

	if *t < 0 || *t > 10 {
		*t = 3
	}

	cookie, sleep = *c, time.Duration(*t)

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

	var wg sync.WaitGroup
	wg.Add(len(p))
	for _, project := range p {
		go func(project string) {
			defer wg.Done()
			err = forceSync(project)
			if err != nil {
				fmt.Printf("[%v] 同步失败\n", project)
				return
			}
		}(project)
	}
	wg.Wait()
}
