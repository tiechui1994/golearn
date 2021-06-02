package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	github = "https://github.com"
)

func GetSessionHiden() (timestamp, token, secret string, err error) {
	u := github + "/login"
	header := map[string]string{
		"Accept": "text/html",
	}
	data, err := GET(u, header)
	if err != nil {
		return timestamp, token, secret, err
	}

	str := strings.Replace(string(data), "\n", "", -1)
	retimestamp := regexp.MustCompile(`<input type="hidden" name="timestamp" value="(.*?)"`)
	retoken := regexp.MustCompile(`<input type="hidden" name="authenticity_token" value="(.*?)"`)
	resecret := regexp.MustCompile(`<input type="hidden" name="timestamp_secret" value="(.*?)"`)

	timestamps := retimestamp.FindAllStringSubmatch(str, 1)
	tokens := retoken.FindAllStringSubmatch(str, 1)
	secrets := resecret.FindAllStringSubmatch(str, 1)
	if len(timestamps) >= 1 && len(timestamps[0]) == 2 &&
		len(tokens) >= 1 && len(tokens[0]) == 2 &&
		len(secrets) >= 1 && len(secrets[0]) == 2 {
		return timestamps[0][1], tokens[0][1], secrets[0][1], nil
	}

	return timestamp, token, secret, errors.New("invalid")
}

func Session(username, password string) (err error) {
	u := github + "/session"
	timestamp, token, secret, err := GetSessionHiden()
	if err != nil {
		return err
	}

	w := url.Values{}
	w.Set("commit", "Sign in")
	w.Set("authenticity_token", token)
	w.Set("login", username)
	w.Set("password", password)
	w.Set("trusted_device", "")
	w.Set("webauthn-support", "supported")
	w.Set("webauthn-iuvpaa-support", "unsupported")
	w.Set("return_to", "")
	w.Set("allow_signup", "")
	w.Set("client_id", "")
	w.Set("integration", "")
	w.Set("required_field_04f6", "")
	w.Set("timestamp", timestamp)
	w.Set("timestamp_secret", secret)

	header := map[string]string{
		"content-type": "application/x-www-form-urlencoded",
		"accept":       "text/html",
	}

	_, err = POST(u, bytes.NewBufferString(w.Encode()), header)
	if err == nil {
		CookieSync <- struct{}{}
	}
	return err
}

func RunDelJobs(project string, day int64) {
	if day == 0 {
		day = 3
	}

	type record struct {
		url  string
		time string
	}

	retotal := regexp.MustCompile(`data-test-selector="workflow-results">.*?([0-9,]+)`)
	rerecord := regexp.MustCompile(`id="check_suite_[0-9]+"\s+data-channel=".*?"\s+.*?<time-ago datetime="(.*?)".*?<details-dialog\s+src="(.*?)"`)
	retoken := regexp.MustCompile(`name="authenticity_token"\s+value="(.*?)"`)

	const size = 25
	var (
		page, total int64
		desc        bool
		steup       bool
	)

	page = 1

loop:
	fmt.Printf("project:[%s], page=%d\n", project, page)
	u := github + fmt.Sprintf("/%s/actions?page=%v", project, page)
	data, err := GET(u, nil)
	if err != nil {
		return
	}

	var records []record
	str := strings.Replace(string(data), "\n", "", -1)
	if page == 1 && !steup {
		steup = true
		tokens := retotal.FindAllStringSubmatch(str, 1)
		if len(tokens) == 1 && len(tokens[0]) == 2 {
			num := strings.ReplaceAll(tokens[0][1], ",", "")
			total, _ = strconv.ParseInt(num, 10, 64)
		}
		if total > 0 {
			page = total / size
			if total%size != 0 {
				page += 1
			}
			fmt.Printf("project:[%s], total=%d, pages=%d\n", project, total, page)
			if page > 1 {
				desc = true
				goto loop
			}
		}
	}

	tokens := rerecord.FindAllStringSubmatch(str, -1)
	for _, token := range tokens {
		records = append(records, record{
			url:  token[2],
			time: token[1],
		})
	}

	del := func(u, token string) {
		value := url.Values{}
		value.Set("_method", "delete")
		value.Set("authenticity_token", token)
		header := map[string]string{
			"accept":       "text/html",
			"content-type": "application/x-www-form-urlencoded",
		}
		_, err = POST(github+u, bytes.NewBufferString(value.Encode()), header)
		if err != nil {
			fmt.Println(u, err)
			return
		}

		fmt.Printf("del [%v] success\n", u)
	}

	for i := len(records) - 1; i >= 0; i-- {
		record := records[i]
		t, _ := time.Parse("2006-01-02T15:04:05Z", record.time)
		if t.Unix()+day*24*3600 > time.Now().Unix() {
			if desc {
				return
			}

			continue
		}

		u = github + record.url
		header := map[string]string{
			"accept":           "text/html",
			"x-requested-with": "XMLHttpRequest",
		}
		data, err = GET(u, header)
		if err != nil {
			fmt.Println(record.url, err)
			continue
		}

		tokens := retoken.FindAllStringSubmatch(string(data), 1)
		if len(tokens) == 1 && len(tokens[0]) == 2 {
			u := record.url[:strings.LastIndex(record.url, "/")]
			go del(u, tokens[0][1])
		}
	}

	if desc && page > 1 {
		page -= 1
		goto loop
	} else if !desc && len(records) == size {
		page += 1
		goto loop
	}
}

func main() {
	RunDelJobs("tiechui1994/jobs", 3)
	time.Sleep(30 * time.Second)
}
