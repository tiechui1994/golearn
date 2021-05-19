package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// http://idea.94goo.com/
// http://idea.do198.com/

// go build -o idea -ldflags '-w -s' idea.go

func init() {
	log.SetPrefix("tool ")
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

func getCode(u string) string {
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

	code := getCode("http://idea.94goo.com")
	if code == "" {
		return
	}

	paths := searchFile(*dir)
	for _, path := range paths {
		writeCode(code, path)
	}
}
