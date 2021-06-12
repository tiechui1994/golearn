package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// http://idea.94goo.com/
// http://idea.do198.com/
// http://vrg123.com/loadcode/ 4565

// go build -o idea -ldflags '-w -s' idea.go

func init() {
	log.SetPrefix("idea ")
}

func writeCode(code, file string) {
	if fd, err := os.Open(file); err == nil {
		data, err := ioutil.ReadAll(fd)
		if err != nil {
			log.Println(err)
			return
		}
		prefix := []byte{0xff, 0xff}
		for _, b := range []byte("URL") {
			prefix = append(prefix, []byte{b, 0x00}...)
		}
		if strings.HasPrefix(string(data), string(prefix)) {
			return
		}
	}

	log.Printf("Start write code to file: %v ...", file)
	fd, err := os.Create(file)
	if err != nil {
		log.Println(err)
		return
	}
	defer fd.Close()

	code = "<certificate-key>\n" + code
	data := []byte{0xff, 0xff}
	for _, b := range []byte(code) {
		data = append(data, []byte{b, 0x00}...)
	}
	fd.Write(data)
	log.Printf("Success write %v", file)
}

func getCode1(u string) string {
	data, err := GET(u, nil)
	if err != nil {
		log.Println(err)
		return ""
	}

	re := regexp.MustCompile(`<input type="hidden" class="new-key" value="(.*)">`)
	tokens := re.FindAllStringSubmatch(string(data), 1)
	if len(tokens) == 1 {
		return tokens[0][1]
	}

	return ""
}

func getCode2() string {
	data, err := GET("http://vrg123.com/", nil)
	if err != nil {
		log.Println(err)
		return ""
	}

	str := strings.Replace(string(data),"\n","",-1)
	var token string
	re := regexp.MustCompile(`<input type="hidden" name="csrfmiddlewaretoken" value="(.*?)">`)
	tokens := re.FindAllStringSubmatch(str, 1)
	if len(tokens) == 1 && len(tokens[0]) == 2 {
		token = tokens[0][1]
	}

	if token == "" {
		return ""
	}

	val := url.Values{}
	val.Set("password", "4565")
	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		"Cookie":       "csrftoken=" + token,
		"X-CSRFToken":  token,
	}
	u := "http://vrg123.com/loadcode/"
	data, err = POST(u, bytes.NewBufferString(val.Encode()), header)
	if err != nil {
		log.Println(err)
		return ""
	}

	var result struct {
		Code    int    `json:"code"`
		Codeval string `json:"codeval"`
	}
	json.Unmarshal(data, &result)
	return result.Codeval
}

func validCode(code string) bool {
	tokens := strings.Split(code, "-")
	if len(tokens) < 2 {
		return false
	}

	var meta struct {
		LicenseId string `json:"licenseId"`
		Products  []struct {
			Code     string `json:"code"`
			PaidUpTo string `json:"paidUpTo"`
			Extended bool   `json:"extended"`
		} `json:"products"`
	}

	data, err := base64.StdEncoding.DecodeString(tokens[1])
	if err != nil {
		return false
	}

	err = json.Unmarshal(data, &meta)
	if err != nil {
		return false
	}

	return meta.LicenseId == tokens[0] && len(meta.Products) > 0 &&
		meta.Products[0].PaidUpTo >= time.Now().Format("2006-01-02")
}

func searchFile(dir string) []string {
	var paths []string
	re := regexp.MustCompile(`\.(goland|clion|pycharm|intellijidea)[0-9]{4}\.[0-9]`)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			name := strings.ToLower(info.Name())
			if re.MatchString(name) {
				key := re.FindAllStringSubmatch(name, 1)[0][1]
				switch key {
				case "goland", "clion", "pycharm":
					paths = append(paths, path+"/config/"+key+".key")
				case "intellijidea":
					paths = append(paths, path+"/config/idea.key")
				}
			}
		}

		return nil
	})

	return paths
}

func main() {
	dir := flag.String("path", "/root", "jetbrains work dir")
	flag.Parse()

	code := getCode1("http://idea.94goo.com")
	if !validCode(code) {
		code = getCode2()
	}
	if !validCode(code) {
		log.Println("no code")
		return
	}
	
	paths := searchFile(*dir)
	for _, path := range paths {
		writeCode(code, path)
	}
}
