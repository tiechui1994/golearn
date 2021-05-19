package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	tcs = "https://tcs.teambition.net"
	www = "https://www.teambition.com"
)

var (
	escape = func() func(name string) string {
		replace := strings.NewReplacer("\\", "\\\\", `"`, "\\\"")
		return func(name string) string {
			return replace.Replace(name)
		}
	}()
)

var cookies = "TEAMBITION_SESSIONID=xxx;TEAMBITION_SESSIONID.sig=xxx;TB_ACCESS_TOKEN=xxx"

type UploadInfo struct {
	FileKey      string `json:"fileKey"`
	FileName     string `json:"fileTame"`
	FileType     string `json:"fileType"`
	FileSize     int    `json:"fileSize"`
	FileCategory string `json:"fileCategory"`
	Source       string `json:"source"`
}

func Upload(token, path string) (url string, err error) {
	fd, err := os.Open(path)
	if err != nil {
		return url, err
	}

	info, _ := fd.Stat()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	w.WriteField("name", info.Name())
	w.WriteField("type", "application/octet-stream")
	w.WriteField("size", fmt.Sprintf("%v", info.Size()))
	w.WriteField("lastModifiedDate", info.ModTime().Format("Mon, 02 Jan 2006 15:04:05 GMT+0800 (China Standard Time)"))

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escape("file"), escape(info.Name())))
	h.Set("Content-Type", "application/octet-stream")
	writer, _ := w.CreatePart(h)
	io.Copy(writer, fd)

	w.Close()

	u := tcs + "/upload"
	header := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  w.FormDataContentType(),
	}

	data, err := POST(u, &body, header)
	if err != nil {
		return url, err
	}

	var result struct {
		ChunkSize   int    `json:"chunkSize"`
		Chunks      int    `json:"chunks"`
		FileKey     string `json:"fileKey"`
		FileName    string `json:"fileName"`
		FileSize    int    `json:"fileSize"`
		Storage     string `json:"storage"`
		DownloadUrl string `json:"downloadUrl"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return url, err
	}

	return result.DownloadUrl, err
}

func UploadChunk(token, path string) (url string, err error) {
	fd, err := os.Open(path)
	if err != nil {
		return url, err
	}

	info, _ := fd.Stat()

	u := tcs + "/upload/chunk"
	header := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}
	bin := fmt.Sprintf(`{"fileName":"%v","fileSize":%v,"lastUpdated":"%v"}`,
		info.Name(), info.Size(), info.ModTime().Format("2006-01-02T15:04:05.00Z"))
	fmt.Println(bin)
	data, err := POST(u, bytes.NewBufferString(bin), header)
	if err != nil {
		return url, err
	}

	fmt.Println("chunk", string(data), err)

	var result struct {
		ChunkSize   int    `json:"chunkSize"`
		Chunks      int    `json:"chunks"`
		FileKey     string `json:"fileKey"`
		FileName    string `json:"fileName"`
		FileSize    int    `json:"fileSize"`
		Storage     string `json:"storage"`
		DownloadUrl string `json:"downloadUrl"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return url, err
	}

	var wg sync.WaitGroup
	wg.Add(result.Chunks)
	for i := 1; i <= result.Chunks; i++ {
		idx := i
		go func(idx int) {
			defer wg.Done()
			data := make([]byte, result.ChunkSize)
			n, _ := fd.ReadAt(data, int64((idx-1)*result.ChunkSize))
			data = data[:n]

			u := tcs + fmt.Sprintf("/upload/chunk/%v?chunk=%v&chunks=%v", result.FileKey, idx, result.Chunks)
			header := map[string]string{
				"Authorization": "Bearer " + token,
				"Content-Type":  "application/octet-stream",
			}
			data, err = POST(u, bytes.NewBuffer(data), header)
			if err != nil {
				fmt.Println("chunk", idx, err)
			}
		}(idx)
	}

	wg.Wait()

	u = tcs + fmt.Sprintf("/upload/chunk/%v", result.FileKey)
	header = map[string]string{
		"Content-Length": "0",
		"Authorization":  "Bearer " + token,
		"Content-Type":   "application/json",
	}
	data, err = POST(u, nil, header)

	fmt.Println("merge", string(data), err)

	if err != nil {
		return url, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return url, err
	}

	return result.DownloadUrl, err
}

type Project struct {
	ID               string `json:"_id"`
	Name             string `json:"name"`
	OrganizationId   string `json:"_organizationId"`
	RootCollectionId string `json:"_rootCollectionId"`
}

func Projects(orgid string) (list []Project, err error) {
	ts := 1621340651679
	u := www + fmt.Sprintf("/api/v2/projects?_organizationId=%v&selectBy=joined&orderBy=name&pageToken=&pageSize=20&_=%v",
		orgid, ts)
	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}
	data, err := GET(u, header)
	if err != nil {
		return list, err
	}

	var result struct {
		Result []Project `json:"result"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return list, err
	}

	return result.Result, nil
}

type Collection struct {
	ID              string `json:"_id"`
	Pinyin          string `json:"pinyin"`
	Title           string `json:"title"`
	ParentId        string `json:"_parentId"`
	ProjectId       string `json:"_projectId"`
	ObjectType      string `json:"objectType"`
	CollectionCount int    `json:"collectionCount"`
	WorkCount       int    `json:"workCount"`
}

type Work struct {
	FileKey     string `json:"fileKey"`
	FileName    string `json:"fileName"`
	DownloadUrl string `json:"downloadUrl"`
	ProjectId   string `json:"_projectId"`
	ParentId    string `json:"_parentId"`
	ObjectType  string `json:"objectType"`
}

func Collections(rootcollid, projectid string) (list []Collection, err error) {
	ts := 1621340651679
	u := www + fmt.Sprintf("/api/collections?_parentId=%v&_projectId=%v&order=updatedDesc&count=50&page=1&_=%v",
		rootcollid, projectid, ts)
	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}
	data, err := GET(u, header)
	if err != nil {
		return list, err
	}

	err = json.Unmarshal(data, &list)
	if err != nil {
		return list, err
	}

	return list, nil
}

func Works(rootcollid, projectid string) (list []Work, err error) {
	ts := 1621340651679
	u := www + fmt.Sprintf("/api/works?_parentId=%v&_projectId=%v&order=updatedDesc&count=50&page=1&_=%v",
		rootcollid, projectid, ts)
	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}
	data, err := GET(u, header)
	if err != nil {
		return list, err
	}

	err = json.Unmarshal(data, &list)
	if err != nil {
		return list, err
	}

	return list, nil
}

func CreateWork(parentid string, upload UploadInfo) error {
	type file struct {
		UploadInfo
		InvolveMembers []interface{} `json:"involveMembers"`
		Visible        string        `json:"visible"`
		ParentId       string        `json:"_parentId"`
	}

	var body struct {
		Works    []file `json:"works"`
		ParentId string `json:"_parentId"`
	}

	f := file{
		UploadInfo: upload,
		Visible:    "members",
		ParentId:   parentid,
	}

	body.Works = []file{f}
	body.ParentId = parentid
	bin, _ := json.Marshal(body)

	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}

	u := www + "/api/works"

	_, err := POST(u, bytes.NewBuffer(bin), header)

	return err
}

func CreateCollection(parentid, projectid, name string) error {
	var body struct {
		CollectionType string        `json:"collectionType"`
		Color          string        `json:"color"`
		Created        string        `json:"created"`
		Description    string        `json:"description"`
		ObjectType     string        `json:"objectType"`
		RecentWorks    []interface{} `json:"recentWorks"`
		SubCount       interface{}   `json:"subCount"`
		Title          string        `json:"title"`
		Updated        string        `json:"updated"`
		WorkCount      int           `json:"workCount"`
		CreatorId      string        `json:"_creatorId"`
		ParentId       string        `json:"_parentId"`
		ProjectId      string        `json:"_projectId"`
	}

	body.Color = "blue"
	body.ObjectType = "collection"
	body.Title = name
	body.ParentId = parentid
	body.ProjectId = projectid
	bin, _ := json.Marshal(body)

	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}

	u := www + "/api/collections"

	_, err := POST(u, bytes.NewBuffer(bin), header)

	return err
}

func FindDir(dir, orgid string) (collection Collection, err error) {
	dir = strings.TrimSpace(dir)
	if !strings.HasPrefix(dir, "/") {
		return collection, errors.New("invalid path")
	}

	tokens := strings.Split(dir[1:], "/")
	projects, err := Projects(orgid)
	if err != nil {
		return collection, err
	}

	var project Project
	exist := false
	for _, p := range projects {
		if p.Name == tokens[0] {
			exist = true
			project = p
			break
		}
	}

	if !exist {
		return collection, errors.New("no exist project")
	}

	tokens = tokens[1:]
	rootid := project.RootCollectionId
	for _, token := range tokens {
		collections, err := Collections(rootid, project.ID)
		if err != nil {
			return collection, err
		}

		exist := false
		for _, c := range collections {
			if c.Title == token {
				collection = c
				rootid = c.ID
				exist = true
				break
			}
		}

		if !exist {
			return collection, errors.New("no exist path: " + token)
		}
	}

	return
}

func FindFile(path, orgid string) (work Work, err error) {
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/") {
		return work, errors.New("invalid path")
	}

	dir, name := filepath.Split(path)
	dir = dir[:len(dir)-1]
	c, err := FindDir(dir, orgid)
	if err != nil {
		return work, err
	}

	works, err := Works(c.ID, c.ProjectId)
	if err != nil {
		return work, err
	}

	for _, work := range works {
		if work.FileName == name {
			return work, nil
		}
	}

	return work, errors.New("not exist file:" + name)
}

func main() {
	dir, err := FindFile("/data/books/wwe/VeePN_2.1.4.0.zip", "000000000000000000000405")
	fmt.Println(dir, err)
}
