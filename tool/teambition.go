package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"mime/multipart"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const (
	tcs = "https://tcs.teambition.net"
	www = "https://www.teambition.com"
	pan = "https://pan.teambition.com"
)

var (
	escape = func() func(name string) string {
		replace := strings.NewReplacer("\\", "\\\\", `"`, "\\\"")
		return func(name string) string {
			return replace.Replace(name)
		}
	}()
	ext = map[string]string{
		".bmp":  "image/bmp",
		".gif":  "image/gif",
		".ico":  "image/vnd.microsoft.icon",
		".jpeg": "image/jpeg",
		".jpg":  "image/jpeg",
		".png":  "image/png",
		".svg":  "image/svg+xml",
		".tif":  "image/tiff",
		".webp": "image/webp",

		".bz":  "application/x-bzip",
		".bz2": "application/x-bzip2",
		".gz":  "application/gzip",
		".rar": "application/vnd.rar",
		".tar": "application/x-tar",
		".zip": "application/zip",
		".7z":  "application/x-7z-compressed",

		".sh":  "application/x-sh",
		".jar": "application/java-archive",
		".pdf": "application/pdf",

		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".xml":  "application/xml",

		".3gp":  "audio/3gpp",
		".3g2":  "audio/3gpp2",
		".wav":  "audio/wav",
		".weba": "audio/webm",
		".oga":  "audio/ogg",
		".mp3":  "audio/mpeg",
		".aac":  "audio/aac",

		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mpeg": "video/mpeg",
		".webm": "video/webm",

		".htm":  "text/html",
		".html": "text/html",
		".js":   "text/javascript",
		".json": "application/json",
		".txt":  "text/plain",
		".text": "text/plain",
		".key":  "text/plain",
		".pem":  "text/plain",
		".cert": "text/plain",
		".csr":  "text/plain",
		".cfg":  "text/plain",
		".go":   "text/plain",
		".java": "text/plain",
		".yml":  "text/plain",
		".md":   "text/plain",
		".s":    "text/plain",
		".c":    "text/plain",
		".cpp":  "text/plain",
		".h":    "text/plain",
		".bin":  "application/octet-stream",
	}
	extType = func() func(string) string {
		return func(s string) string {
			if val, ok := ext[s]; ok {
				return val
			}
			return "application/octet-stream"
		}
	}()

	cookies = "TEAMBITION_SESSIONID=xxx;TEAMBITION_SESSIONID.sig=xxx;TB_ACCESS_TOKEN=xxx"
)

// login
// "pkcs1_oaep"
type OAEP struct {
	alg    string
	hash   hash.Hash
	pubkey string
	prikey string
}

func (o *OAEP) init() {
	switch o.alg {
	case "sha1":
		o.hash = sha1.New()
	case "sha256":
		o.hash = sha256.New()
	case "md5":
		o.hash = md5.New()
	}
}

func (o *OAEP) Encrypt(msg []byte) string {
	block, _ := pem.Decode([]byte(o.pubkey))
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println("Load public key error")
		panic(err)
	}

	encrypted, err := rsa.EncryptOAEP(o.hash, rand.Reader, pub.(*rsa.PublicKey), msg, nil)
	if err != nil {
		fmt.Println("Encrypt data error")
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(encrypted)
}

func (o *OAEP) Decrypt(encrypted string) []byte {
	block, _ := pem.Decode([]byte(o.prikey))
	var pri *rsa.PrivateKey
	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("Load private key error")
		panic(err)
	}

	decodedData, err := base64.StdEncoding.DecodeString(encrypted)
	ciphertext, err := rsa.DecryptOAEP(o.hash, rand.Reader, pri, decodedData, nil)
	if err != nil {
		fmt.Println("Decrypt data error")
		panic(err)
	}

	return ciphertext
}

func Login(clientid, pubkey, token, email, phone, pwd string) (access string, err error) {
	key, _ := base64.StdEncoding.DecodeString(pubkey)
	oeap := OAEP{
		pubkey: string(key),
		alg:    "sha1",
	}
	oeap.init()

	password := oeap.Encrypt([]byte(pwd))

	body := map[string]string{
		"password":      password,
		"client_id":     clientid,
		"response_type": "session",
		"publicKey":     pubkey,
		"token":         token,
	}

	var u string
	if email != "" {
		u = "/api/login/email"
		body["email"] = email
	} else {
		u = "/api/login/phone"
		body["phone"] = phone
	}

	u = "https://account.teambition.com" + u
	var writer bytes.Buffer
	json.NewEncoder(&writer).Encode(body)
	header := map[string]string{
		"content-type": "application/json",
	}
	fmt.Println(writer.String())
	data, err := POST(u, &writer, header)
	if err != nil {
		if _, ok := err.(CodeError); ok {
			log.Println("login", string(data))
		}
		return access, err
	}

	var result struct {
		AbnormalLogin      string `json:"abnormalLogin"`
		HasGoogleTwoFactor bool   `json:"hasGoogleTwoFactor"`
		Token              string `json:"token"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return access, err
	}

	if result.HasGoogleTwoFactor {
		err = TwoFactor(clientid, token, result.Token)
		if err != nil {
			return access, err
		}
	}

	if jar != nil {
		cook := jar.(*cookiejar.Jar)
		localjar := (*Jar)(unsafe.Pointer(cook))
		teambit := localjar.Entries["teambition.com"]
		teambit["teambition.com;/;TB_ACCESS_TOKEN"] = entry{
			Name:       "TB_ACCESS_TOKEN",
			Value:      result.Token,
			Domain:     "teambition.com",
			Path:       "/",
			Secure:     true,
			HttpOnly:   true,
			Persistent: true,
			HostOnly:   false,
			Expires:    time.Now().Add(30 * time.Hour * 24),
			Creation:   time.Now(),
			SeqNum:     4,
		}
		localjar.Entries["teambition.com"] = teambit
	}

	return result.Token, nil
}

func TwoFactor(clientid, token, verify string) error {
	var code string
	fmt.Printf("Input Auth Code:")
	fmt.Scanf("%v", &code)

	body := map[string]string{
		"authcode":      code,
		"client_id":     clientid,
		"response_type": "session",
		"token":         token,
		"verify":        verify,
	}

	u := "https://account.teambition.com/api/login/two-factor"
	var writer bytes.Buffer
	json.NewEncoder(&writer).Encode(body)
	header := map[string]string{
		"content-type": "application/json",
	}
	data, err := POST(u, &writer, header)
	if err != nil {
		if _, ok := err.(CodeError); ok {
			log.Println("two-factor", string(data))
			return err
		}

		return err
	}

	return nil
}

func LoginParams() (clientid, token, publickey string, err error) {
	raw, err := GET("https://account.teambition.com/login", nil)
	if err != nil {
		return clientid, token, publickey, err
	}

	reaccout := regexp.MustCompile(`<script id="accounts-config" type="application/json">(.*?)</script>`)
	republic := regexp.MustCompile(`<script id="accounts-ssr-props" type="application/react-ssr-props">(.*?)</script>`)

	str := strings.Replace(string(raw), "\n", "", -1)
	rawa := reaccout.FindAllStringSubmatch(str, 1)
	rawp := republic.FindAllStringSubmatch(str, 1)
	if len(rawa) > 0 && len(rawa[0]) == 2 && len(rawp) > 0 && len(rawp[0]) == 2 {
		var config struct {
			TOKEN     string
			CLIENT_ID string
		}
		err = json.Unmarshal([]byte(rawa[0][1]), &config)
		if err != nil {
			return
		}

		var public struct {
			Fsm struct {
				Config struct {
					Pub struct {
						Algorithm string `json:"algorithm"`
						PublicKey string `json:"publicKey"`
					} `json:"pub"`
				} `json:"config"`
			} `json:"fsm"`
		}

		pub, _ := url.QueryUnescape(rawp[0][1])
		err = json.Unmarshal([]byte(pub), &public)
		if err != nil {
			return
		}

		return config.CLIENT_ID, config.TOKEN, public.Fsm.Config.Pub.PublicKey, nil
	}

	return clientid, token, publickey, errors.New("api change update")
}

// project
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
	w.WriteField("type", extType(filepath.Ext(info.Name())))
	w.WriteField("size", fmt.Sprintf("%v", info.Size()))
	w.WriteField("lastModifiedDate", info.ModTime().Format("Mon, 02 Jan 2006 15:04:05 GMT+0800 (China Standard Time)"))

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escape("file"), escape(info.Name())))
	h.Set("Content-Type", extType(filepath.Ext(info.Name())))
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
	ts := time.Now().UnixNano() / 1e6
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
	ts := time.Now().UnixNano() / 1e6
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
	ts := time.Now().UnixNano() / 1e6
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

type MeConfig struct {
	IsAdmin bool   `json:"isAdmin"`
	OrgId   string `json:"tenantId"`
	User    struct {
		ID     string `json:"id"`
		Email  string `json:"email"`
		Name   string `json:"name"`
		OpenId string `json:"openId"`
	}
}

func Batches() (me MeConfig, err error) {
	u := www + "/uiless/api/sdk/batch?scope[]=me"
	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}
	data, err := GET(u, header)
	if err != nil {
		return me, err
	}
	var result struct {
		Result struct {
			Me MeConfig `json:"me"`
		} `json:"result"`
	}

	err = json.Unmarshal(data, &result)
	return result.Result.Me, err
}

// netdisk

type Role struct {
	ID             string   `json:"_id"`
	OrganizationId string   `json:"_organizationId"`
	Level          int      `json:"level"`
	Permissions    []string `json:"-"`
}

func Roles() (list []Role, err error) {
	ts := time.Now().UnixNano() / 1e6
	u := www + fmt.Sprintf("/api/v2/roles?type=organization&_=%v", ts)
	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}
	data, err := GET(u, header)
	if err != nil {
		return list, err
	}

	var result struct {
		Result struct {
			Roles []Role `json:"roles"`
		} `json:"result"`
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return list, err
	}

	return result.Result.Roles, nil
}

// 文件上传: CreateFolder -> CreateFile -> UploadUrl -> UploadPanFile
type Node struct {
	Kind        string `json:"kind"` // file, folder
	Name        string `json:"name"`
	ParentId    string `json:"parentId"`
	NodeId      string `json:"nodeId"`
	Status      string `json:"status"`
	DriveId     string `json:"driveId"`
	ContainerId string `json:"containerId"`

	DownloadUrl string `json:"downloadUrl"`
	Url         string `json:"url"`
}

func Nodes(orgid, driveid, parentid string) (list []Node, err error) {
	u := pan + fmt.Sprintf("/pan/api/nodes?orgId=%v&from=&limit=100&orderBy=updated_at&orderDirection=DESC&driveId=%v&parentId=%v",
		orgid, driveid, parentid)
	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}
	data, err := GET(u, header)
	if err != nil {
		return list, err
	}

	var result struct {
		Data []Node `json:"data"`
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return list, err
	}

	return result.Data, nil
}

func NodeArchive(orgid string, nodeids []string) (err error) {
	u := pan + "/pan/api/nodes/archive"

	var body struct {
		NodeIds []string `json:"nodeIds"`
		OrgId   string   `json:"orgId"`
	}
	body.NodeIds = nodeids
	body.OrgId = orgid
	bin, _ := json.Marshal(body)

	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}

	_, err = POST(u, bytes.NewBuffer(bin), header)
	return err
}

func NodeMove(orgid, driverid string, nodeids []string, dstparentid string) (err error) {
	u := pan + "/pan/api/nodes/move"

	var body struct {
		DriveId   string              `json:"driveId"`
		Ids       []map[string]string `json:"ids"`
		OrgId     string              `json:"orgId"`
		ParentId  string              `json:"parentId"`
		SameLevel bool                `json:"sameLevel"`
	}
	body.DriveId = driverid
	for _, id := range nodeids {
		body.Ids = append(body.Ids, map[string]string{
			"id":        id,
			"ccpFileId": id,
		})
	}
	body.ParentId = dstparentid
	body.OrgId = orgid
	bin, _ := json.Marshal(body)

	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}

	_, err = POST(u, bytes.NewBuffer(bin), header)
	return err
}

func CreateFolder(name, orgid, parentid, spaceid, driverid string) (nodeid string, err error) {
	var body struct {
		CcpParentId   string `json:"ccpParentId"`
		CheckNameMode string `json:"checkNameMode"`
		DriveId       string `json:"driveId"`
		Name          string `json:"name"`
		OrgId         string `json:"orgId"`
		ParentId      string `json:"parentId"`
		SpaceId       string `json:"spaceId"`
		Type          string `json:"type"`
	}

	body.CheckNameMode = "refuse"
	body.DriveId = driverid
	body.Name = name
	body.OrgId = orgid
	body.CcpParentId = parentid
	body.ParentId = parentid
	body.SpaceId = spaceid
	body.Type = "folder"
	bin, _ := json.Marshal(body)

	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}

	u := pan + "/pan/api/nodes/folder"

	data, err := POST(u, bytes.NewBuffer(bin), header)
	if err != nil {
		return nodeid, err
	}

	var result []struct {
		NodeId string `json:"nodeId"`
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nodeid, err
	}

	return result[0].NodeId, nil
}

type PanFile struct {
	OrgId           string   `json:"orgId"`
	Name            string   `json:"name"`
	Kind            string   `json:"kind"`
	UploadId        string   `json:"uploadId"`
	UploadUrl       []string `json:"uploadUrl"`
	NodeId          string   `json:"nodeId"`
	ParentId        string   `json:"parentId"`
	DriveId         string   `json:"driveId"`
	CcpFileId       string   `json:"ccpFileId"`       // NodeId
	CcpParentFileId string   `json:"ccpParentFileId"` // ParentID
}

func CreateFile(orgid, parentid, spaceid, driverid, path string) (files []PanFile, err error) {
	type file struct {
		Name        string `json:"name"`
		ContentType string `json:"contentType"`
		ChunkCount  int    `json:"chunkCount"`
		Size        int64  `json:"size"`
		CcpParentId string `json:"ccpParentId"`
		DriveId     string `json:"driveId"`
		Type        string `json:"type"`
	}
	var body struct {
		CheckNameMode string `json:"checkNameMode"`
		Infos         []file `json:"infos"`
		OrgId         string `json:"orgId"`
		ParentId      string `json:"parentId"`
		SpaceId       string `json:"spaceId"`
	}

	info, _ := os.Stat(path)

	body.Infos = []file{{
		Name:        info.Name(),
		ChunkCount:  1,
		ContentType: extType(filepath.Ext(info.Name())),
		DriveId:     driverid,
		Size:        info.Size(),
		CcpParentId: parentid,
		Type:        "file",
	}}
	body.CheckNameMode = "refuse"
	body.OrgId = orgid
	body.ParentId = parentid
	body.SpaceId = spaceid
	bin, _ := json.Marshal(body)

	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}

	u := pan + "/pan/api/nodes/file"

	data, err := POST(u, bytes.NewBuffer(bin), header)
	if err != nil {
		return files, err
	}

	err = json.Unmarshal(data, &files)
	return files, nil
}

type PanUpload struct {
	DomainId     string `json:"domainId"`
	DriveId      string `json:"driveId"`
	UploadId     string `json:"uploadId"`
	FileId       string `json:"fileId"`
	NodeId       string `json:"nodeid"`
	PartInfoList []struct {
		PartNumber int    `json:"partNumber"`
		UploadUrl  string `json:"uploadUrl"`
	} `json:"partInfoList"`
}

func UploadUrl(orgid string, panfile PanFile) (upload PanUpload, err error) {
	var body struct {
		DriveId         string `json:"driveId"`
		OrgId           string `json:"orgId"`
		UploadId        string `json:"uploadId"`
		StartPartNumber int    `json:"startPartNumber"`
		EndPartNumber   int    `json:"endPartNumber"`
	}

	body.DriveId = panfile.DriveId
	body.OrgId = orgid
	body.UploadId = panfile.UploadId
	body.StartPartNumber = 1
	body.EndPartNumber = len(panfile.UploadUrl)
	bin, _ := json.Marshal(body)

	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}

	u := pan + fmt.Sprintf("/pan/api/nodes/%v/uploadUrl", panfile.NodeId)

	data, err := POST(u, bytes.NewBuffer(bin), header)
	if err != nil {
		return upload, err
	}

	err = json.Unmarshal(data, &upload)
	upload.NodeId = panfile.NodeId
	return upload, err
}

func UploadPanFile(orgid string, upload PanUpload, path string) error {
	fd, err := os.Open(path)
	if err != nil {
		return err
	}

	info, _ := fd.Stat()
	chunk := uint64(info.Size() / int64(len(upload.PartInfoList)))
	k8 := uint64(8192)
	chunk = k8 + (^(k8 - 1))&chunk /// 8192对其

	var wg sync.WaitGroup
	wg.Add(len(upload.PartInfoList))
	for _, v := range upload.PartInfoList {
		part := v
		go func(u string, partnum int) {
			defer wg.Done()
			reader := make([]byte, chunk)
			fd.ReadAt(reader, int64(partnum-1)*int64(chunk))
			_, err = PUT(u, bytes.NewBuffer(reader), nil)
		}(part.UploadUrl, part.PartNumber)
	}
	wg.Wait()

	var body struct {
		CcpFileId string `json:"ccpFileId"`
		DriveId   string `json:"driveId"`
		NodeId    string `json:"nodeId"`
		OrgId     string `json:"orgId"`
		UploadId  string `json:"uploadId"`
	}
	body.CcpFileId = upload.NodeId
	body.NodeId = upload.NodeId
	body.OrgId = orgid
	body.DriveId = upload.DriveId
	body.UploadId = upload.UploadId
	bin, _ := json.Marshal(body)

	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}

	u := pan + "/pan/api/nodes/complete"

	_, err = POST(u, bytes.NewBuffer(bin), header)
	return err
}

type Org struct {
	OrganizationId string `json:"organizationId"`
	DriveId        string `json:"driveId"`
	AppId          string `json:"_appId"`
	IsPersonal     bool   `json:"isPersonal"`
	IsPublic       bool   `json:"isPublic"`
	Name           string `json:"name"`
	TotalSize      int64  `json:"totalSize"`
	UsedSize       int64  `json:"usedSize"`
}

func Orgs(orgid string) (org Org, err error) {
	u := pan + fmt.Sprintf("/pan/api/orgs/%v?orgId=%v", orgid, orgid)
	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}
	data, err := GET(u, header)
	if err != nil {
		return org, err
	}

	var result struct {
		Data Org `json:"data"`
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return org, err
	}

	return result.Data, nil
}

type User struct {
	Email          string `json:"email"`
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	PhoneForLogin  string `json:"phoneForLogin"`
	ID             string `json:"_id"`
	OrganizationId string `json:"_organizationId"`
	Roleid         string `json:"roleid"`
}

func GetByUser(orgid string) (user User, err error) {
	u := pan + fmt.Sprintf("/pan/api/orgs/%v/members/getByUser?orgId=%v", orgid, orgid)
	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}
	data, err := GET(u, header)
	if err != nil {
		return user, err
	}

	var result struct {
		RoleId          string `json:"_roleId"`
		BoundToObjectId string `json:"_boundToObjectId"`
		UserInfo        User   `json:"userInfo"`
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return user, err
	}

	result.UserInfo.OrganizationId = result.BoundToObjectId
	result.UserInfo.Roleid = result.RoleId

	return result.UserInfo, nil
}

type Space struct {
	IsPublic     bool   `json:"isPublic"`
	MembersCount int    `json:"membersCount"`
	Name         string `json:"name"`
	RootId       string `json:"rootId"`
	SpaceId      string `json:"spaceId"`
}

func Spaces(orgid, memberid string) (spaces []Space, err error) {
	u := pan + fmt.Sprintf("/pan/api/spaces?orgId=%v&memberId=%v", orgid, memberid)
	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}
	data, err := GET(u, header)
	if err != nil {
		return spaces, err
	}

	err = json.Unmarshal(data, &spaces)
	if err != nil {
		return spaces, err
	}

	return spaces, nil
}

func GetCacheData() (roles []Role, org Org, spaces []Space, err error) {
	var result struct {
		Roles  []Role
		Org    Org
		Spaces []Space
	}
	if ReadFile("/tmp/teambit.json", &result) == nil {
		return result.Roles, result.Org, result.Spaces, nil
	}

	roles, err = Roles()
	if err != nil {
		log.Println(err)
		return
	}

	if len(roles) == 0 {
		err = errors.New("no roles")
		return
	}

	org, err = Orgs(roles[0].OrganizationId)
	if err != nil {
		log.Println(err)
		return
	}

	var user User
	user, err = GetByUser(roles[0].OrganizationId)
	if err != nil {
		log.Println(err)
		return
	}

	spaces, err = Spaces(user.OrganizationId, user.ID)
	if err != nil {
		log.Println(err)
		return
	}

	if len(spaces) == 0 {
		err = errors.New("no spaces")
		return
	}

	result.Roles = roles
	result.Org = org
	result.Spaces = spaces
	WriteFile("/tmp/teambit.json", result)

	return
}

func PanFindDir(dir string) (node Node, err error) {
	dir = strings.TrimSpace(dir)
	if !strings.HasPrefix(dir, "/") {
		return node, errors.New("invalid path")
	}

	_, org, spaces, err := GetCacheData()
	if err != nil {
		return node, err
	}

	tokens := strings.Split(dir[1:], "/")
	parentid := spaces[0].RootId
	exist := false
	for i, p := range tokens {
		nodes, err := Nodes(org.OrganizationId, org.DriveId, parentid)
		if err != nil {
			return node, err
		}

		for _, n := range nodes {
			if n.Name == p {
				parentid = n.NodeId
				node = n
				exist = i == len(tokens)-1
			}
		}
	}

	if !exist {
		return node, errors.New("no exist path")
	}

	return node, nil
}

func PanMkdirP(dir string) (nodeid string, err error) {
	dir = strings.TrimSpace(dir)
	if !strings.HasPrefix(dir, "/") {
		return nodeid, errors.New("invalid path")
	}

	_, org, spaces, err := GetCacheData()
	if err != nil {
		return nodeid, err
	}

	isSearch := true
	tokens := strings.Split(dir[1:], "/")
	parentid := spaces[0].RootId
	for _, p := range tokens {
		if isSearch {
			nodes, err := Nodes(org.OrganizationId, org.DriveId, parentid)
			if err != nil {
				return nodeid, err
			}

			exist := false
			for _, n := range nodes {
				if n.Name == p {
					parentid = n.NodeId
					exist = true
					break
				}
			}

			if !exist {
				isSearch = false
				parentid, err = CreateFolder(p, org.OrganizationId, parentid, spaces[0].SpaceId, org.DriveId)
				if err != nil {
					return nodeid, err
				}
			}

			continue
		}

		parentid, err = CreateFolder(p, org.OrganizationId, parentid, spaces[0].SpaceId, org.DriveId)
		if err != nil {
			return nodeid, err
		}
	}

	return parentid, nil
}

func main() {
	DEBUG = true

	_, org, spaces, err := GetCacheData()
	if err != nil {
		log.Println(err)
		return
	}

	nodeid, err := PanMkdirP("/vpn/安卓")
	if err != nil {
		log.Println("PanMkdirP", err)
		return
	}

	var files []string
	filepath.Walk("/home/user/Downloads/6款打包", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	for _, file := range files {
		files, err := CreateFile(org.OrganizationId, nodeid, spaces[0].SpaceId, org.DriveId, file)
		if err == nil {
			upload, err := UploadUrl(org.OrganizationId, files[0])
			if err == nil {
				UploadPanFile(org.OrganizationId, upload, file)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)
}
