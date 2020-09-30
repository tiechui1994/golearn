package api

import (
	"time"
	"net/http"
	"net"
	"io/ioutil"
	"regexp"
	"fmt"
	"net/url"
	"bytes"
	"encoding/json"
)

var dns = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Minute)
		},
	},
	Timeout: time.Minute,
}

// ping.cn
func DNS(host string) (ips []string, err error) {
	token, err := token(host)
	if err != nil {
		return ips, err
	}

	var taskid string
	_, taskid, err = check(host, token, "", true)
	for err != nil {
		_, taskid, err = check(host, token, "", true)
	}

	stop := false
	timer := time.NewTimer(30 * time.Second)
	for {
		if stop {
			break
		}
		select {
		case <-timer.C:
			stop = true
		default:
			ips, _, err = check(host, token, taskid, false)
			time.Sleep(500 * time.Millisecond)
		}
	}

	return ips, err
}

func token(host string) (string, error) {
	u := "https://www.ping.cn/dns/" + host
	response, err := dns.Get(u)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)
	regex := regexp.MustCompile(`<input\s+value="([0-9a-zA-Z]+)".*name="_token">`)
	tokens := regex.FindStringSubmatch(string(data))
	if len(tokens) >= 2 {
		return tokens[1], nil
	}
	return "", fmt.Errorf("invalid token")
}

func check(host, token, taskid string, isCreate bool) (ips []string, task string, err error) {
	value := url.Values{}
	value.Set("host", host)
	value.Set("_token", token)
	value.Set("type", "dns")

	if isCreate {
		value.Set("create_task", "1")
		value.Set("host2", "")
		value.Set("node_ids", "")
		value.Set("isp", "1,2,3,9,10,11")
		value.Set("dns_server", "")
		value.Set("dns_type", "A")
	} else {
		value.Set("create_task", "0")
		value.Set("task_id", taskid)
	}

	u := "https://www.ping.cn/check"
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(value.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	response, err := dns.Do(request)
	if err != nil {
		return ips, task, err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ips, task, err
	}

	if isCreate {
		var result struct {
			Code int `json:"code"`
			Data struct {
				TaskID string `json:"taskID"`
			} `json:"data"`
		}
		json.Unmarshal(data, &result)
		if result.Code == 1 {
			return ips, result.Data.TaskID, nil
		}

		return ips, "", fmt.Errorf("code:%v", result.Code)
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			InitData struct {
				IpPre []struct {
					IP  string  `json:"ip"`
					Pre float64 `json:"pre"`
				}
			}
		} `json:"data"`
	}

	json.Unmarshal(data, &result)
	if result.Code != 1 {
		return ips, "", fmt.Errorf("code:%v", result.Code)
	}

	for _, ip := range result.Data.InitData.IpPre {
		ips = append(ips, ip.IP)
	}

	return ips, taskid, nil
}
