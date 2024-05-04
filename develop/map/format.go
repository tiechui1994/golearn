package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)


func main() {
	md := flag.String("md", "", "markdown file or dir")
	flag.Parse()
	if *md == "" {
		return
	}

	stat, err := os.Stat(*md)
	if err != nil {
		return
	}

	files := []string{*md}
	if stat.IsDir() {
		files = files[:0]
		_ = filepath.Walk(*md, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				files = append(files, path)
			}

			return nil
		})
	}


	for _, file := range files {
		all, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println("ReadFile", file, "failed:", err)
			return
		}

		reg := regexp.MustCompile(`(\t+)`)
		lines := strings.Split(string(all), "\n")
		var start bool
		for i, v := range lines {
			vspace := strings.TrimSpace(v)
			if len(vspace) == 0 {
				continue
			}

			if strings.HasPrefix(vspace, "```") && start {
				start = false
				continue
			}

			if strings.HasPrefix(vspace, "```") && !start {
				start = true
				continue
			}

			if start {
				if reg.MatchString(v) {
					tokens := reg.FindAllStringSubmatch(v, -1)
					c := strings.Count(tokens[0][0], "\t")
					lines[i] = strings.ReplaceAll(v,
						strings.Repeat("\t", c),
						strings.Repeat("    ", c))
				}
			}
		}

		_ = ioutil.WriteFile(file, []byte(strings.Join(lines, "\n")), 0666)
	}
}
