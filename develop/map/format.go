package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)


func main() {
	md := flag.String("md", "", "markdown file")
	flag.Parse()
	if *md == "" {
		return
	}

	all, err := ioutil.ReadFile(*md)
	if err != nil {
		fmt.Println("ReadFile", *md, "failed:", err)
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

	ioutil.WriteFile(*md, []byte(strings.Join(lines, "\n")), 0666)
}
