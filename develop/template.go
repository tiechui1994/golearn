package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

func List() {
	type FileNode struct {
		Name    string
		Updated string
		Type    int
	}

	tpl := `
total {{ (len $) }}
{{- range $ }}
{{- $x := "f" }}
{{- if (eq .Type 1) }} 
	{{- $x = "d" }}
{{- end }}
{{printf "%s  %24s  %s" $x .Updated .Name}}
{{- end }}
`
	tpl = strings.Trim(tpl, "\n")
	temp, err := template.New("").Parse(tpl)
	if err != nil {
		fmt.Println(err)
		return
	}

	list := []FileNode{
		{
			Name:    "111",
			Updated: "2021-04-24T07:36:48.591Z",
			Type:    1,
		},
		{
			Name:    "222",
			Updated: "2021-04-24T07:36:48.591Z",
			Type:    2,
		},
		{
			Name:    "333333",
			Updated: "2021-04-24T07:36:48.591Z",
			Type:    2,
		},
	}

	var buf bytes.Buffer
	err = temp.Execute(&buf, list)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("=======================")
	fmt.Println(buf.String())
}

func Main() {
	tpl := `
{{ $ }}
{{ $x := (len $) -}}
{{ println $x }}
`
	tpl = strings.Trim(tpl, "\n")
	temp, err := template.New("T1").Parse(tpl)
	if err != nil {
		fmt.Println("Parse", err)
		return
	}

	data := []string{"AA", "BB"}
	var buf bytes.Buffer
	err = temp.Execute(&buf, data)
	if err != nil {
		fmt.Println("Execute", err)
		return
	}

	fmt.Println(buf.String())
}

func main() {
	Main()
	List()
}
