package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
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

func RunJobs(project string) {
	u := github + "/" + project + "/actions?page=5"
	data, err := GET(u, nil)
	if err != nil {
		return
	}

	type record struct {
		id   string
		time string
	}

	var records []record
	str := strings.Replace(string(data), "\n", "", -1)
	reinfos := regexp.MustCompile(`id="check_suite_[0-9]+"\s+data-channel=".*?"\s+data-url="(.*?)">.*?<time-ago datetime="(.*?)"`)
	reid := regexp.MustCompile("/" + project + "/actions/workflow-run/([0-9]+)")
	tokens := reinfos.FindAllStringSubmatch(str, -1)
	for _, token := range tokens {
		ids := reid.FindAllStringSubmatch(token[1], 1)
		records = append(records, record{
			id:   ids[0][1],
			time: token[1],
		})
	}

	//rehiden := regexp.MustCompile(`name="authenticity_token" value="(.*?)"`)
	for _, record := range records {
		u = github + "/" + project + "/actions/runs/" + record.id + "/delete"
		u = "https://github.com/tiechui1994/jobs/actions/runs/879808969/delete"
		fmt.Println(u)
		header := map[string]string{
			"accept": "text/html",
		}
		data, err = GET(u, header)

		fmt.Println(string(data))
		break
	}

	fmt.Println(len(records))
}

func main() {
	RunJobs("tiechui1994/jobs")
}
