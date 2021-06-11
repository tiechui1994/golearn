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
	"html"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
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

	CookieSync <- struct{}{}

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

//===================================== user  =====================================
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

//===================================== project  =====================================
type UploadInfo struct {
	FileKey      string `json:"fileKey"`
	FileName     string `json:"fileName"`
	FileType     string `json:"fileType"`
	FileSize     int    `json:"fileSize"`
	FileCategory string `json:"fileCategory"`
	Source       string `json:"source"`
	DownloadUrl  string `json:"downloadUrl"`
}

func UploadProjectFile(token, path string) (upload UploadInfo, err error) {
	fd, err := os.Open(path)
	if err != nil {
		return upload, err
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
		return upload, err
	}

	err = json.Unmarshal(data, &upload)
	if err != nil {
		return upload, err
	}

	return upload, err
}

func UploadProjectFileChunk(token, path string) (upload UploadInfo, err error) {
	fd, err := os.Open(path)
	if err != nil {
		return upload, err
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
		return upload, err
	}

	fmt.Println("chunk", string(data), err)

	var result struct {
		UploadInfo
		ChunkSize int `json:"chunkSize"`
		Chunks    int `json:"chunks"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return upload, err
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
		return upload, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return upload, err
	}

	upload = result.UploadInfo
	return upload, err
}

func ArchiveProjectDir(token string, nodeid, projectid, name, targetdir string) (err error) {
	u := www + fmt.Sprintf("/api/projects/%v/download-info?_collectionIds=%v&_workIds=&zipName=%v",
		projectid, nodeid, name)
	header := map[string]string{
		"cookie":       cookies,
		"content-type": "application/json; charset=utf-8",
	}

	data, err := GET(u, header)
	if err != nil {
		return err
	}

	type item struct {
		Directories  []item   `json:"directories"`
		DownloadUrls []string `json:"downloadUrls"`
		Name         string   `json:"name"`
	}
	var result struct {
		Directories  []item   `json:"directories"`
		DownloadUrls []string `json:"downloadUrls"`
		ZipName      string   `json:"zipName"`
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}

	value := url.Values{}
	var dfs func(prefix string, it []item)
	dfs = func(prefix string, it []item) {
		for idx, item := range it {
			key := prefix + fmt.Sprintf(`[%d][name]`, idx)
			value.Set(key, item.Name)

			keyprefix := prefix + fmt.Sprintf("[%d][directories]", idx)
			dfs(keyprefix, item.Directories)

			for _, val := range item.DownloadUrls {
				key := prefix + fmt.Sprintf("[%d][downloadUrls][]", idx)
				value.Set(key, val)
			}
		}

	}

	dfs("directories", result.Directories)
	for _, val := range result.DownloadUrls {
		key := "downloadUrls[]"
		value.Set(key, val)
	}
	value.Set("zipName", result.ZipName)

	u = "https://tcs.teambition.net/archive?Signature=" + token
	header = map[string]string{
		"content-type": "application/x-www-form-urlencoded",
	}
	return File(u, "POST", bytes.NewBufferString(value.Encode()), header, filepath.Join(targetdir, name+".zip"))
}

type Project struct {
	ID               string `json:"_id"`
	Name             string `json:"name"`
	OrganizationId   string `json:"_organizationId"`
	RootCollectionId string `json:"_rootCollectionId"`
}

// 注: 这个 orgid 比较特殊, 它是一个特殊的 orgid
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

const (
	Object_Collection = "collection"
	Object_Work       = "work"
)

type Collection struct {
	ID              string `json:"_id"`
	Pinyin          string `json:"pinyin"`
	Title           string `json:"title"`
	ParentId        string `json:"_parentId"`
	ProjectId       string `json:"_projectId"`
	ObjectType      string `json:"objectType"`
	CollectionCount int    `json:"collectionCount"`
	WorkCount       int    `json:"workCount"`
	Updated         string `json:"updated"`
}

type Work struct {
	ID          string `json:"_id"`
	FileKey     string `json:"fileKey"`
	FileName    string `json:"fileName"`
	FileSize    int    `json:"fileSize"`
	DownloadUrl string `json:"downloadUrl"`
	ProjectId   string `json:"_projectId"`
	ParentId    string `json:"_parentId"`
	ObjectType  string `json:"objectType"`
	Updated     string `json:"updated"`
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

func Archive(nodeid string) (err error) {
	body := `{}`
	u := www + "api/works/" + nodeid + "/archive"
	header := map[string]string{
		"content-type": "application/json; charset=utf-8",
	}
	_, err = POST(u, bytes.NewBufferString(body), header)
	return err
}

func FindProjectDir(dir, orgid string) (collection Collection, err error) {
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

func FindProjectFile(path, orgid string) (work Work, err error) {
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/") {
		return work, errors.New("invalid path")
	}

	dir, name := filepath.Split(path)
	dir = dir[:len(dir)-1]
	c, err := FindProjectDir(dir, orgid)
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

func GetProjectToken(projectid, rootcollid string) (token string, err error) {
	u := www + "/project/" + projectid + "/works/" + rootcollid
	header := map[string]string{
		"accept": "text/html",
	}
	data, err := GET(u, header)
	if err != nil {
		return token, err
	}

	str := strings.Replace(string(data), "\n", "", -1)
	reconfig := regexp.MustCompile(`<span\s+id="teambition-config".*?>(.*)</span>`)
	raw := reconfig.FindAllStringSubmatch(str, 1)
	if len(raw) == 1 && len(raw[0]) == 2 {
		var result struct {
			UserInfo struct {
				StrikerAuth string `json:"strikerAuth"`
			} `json:"userInfo"`
		}
		config, _ := url.QueryUnescape(html.UnescapeString(raw[0][1]))
		err = json.Unmarshal([]byte(config), &result)
		if err != nil {
			return token, err
		}

		return result.UserInfo.StrikerAuth[7:], nil
	}

	return token, errors.New("no tokens")
}

//===================================== pan  =====================================

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

func CreatePanFile(orgid, parentid, spaceid, driverid, path string) (files []PanFile, err error) {
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

func CreatePanUpload(orgid string, panfile PanFile) (upload PanUpload, err error) {
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

	key := filepath.Join(ConfDir, "teambition_cache.json")
	if ReadFile(key, &result) == nil {
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
	WriteFile(key, result)

	return
}

type FileSystem interface {
	Init() error
	DownloadUrl(srcpath, targetdir string) error
	UploadFile(filepath, targetdir string) error
	UploadDir(filepath, targetdir string) error
}

const (
	Node_Dir  = 1
	Node_File = 2
)

type FileNode struct {
	Type     int                  `json:"type"`
	Name     string               `json:"name"`
	NodeId   string               `json:"nodeid"`
	ParentId string               `json:"parentid"`
	Updated  string               `json:"updated"`
	Child    map[string]*FileNode `json:"child,omitempty"`
	Url      string               `json:"url,omitempty"`
	Size     int                  `json:"size,omitempty"`
	Private  interface{}          `json:"private,omitempty"`
}

type ProjectFs struct {
	Name  string
	Orgid string

	mux        sync.Mutex
	projectid  string
	rootcollid string
	token      string
	root       *FileNode
}

func (p *ProjectFs) Init() (err error) {
	if p.Name == "" || p.Orgid == "" {
		return
	}

	list, err := Projects(p.Orgid)
	if err != nil {
		fmt.Println(err, p.Orgid)
		return err
	}

	for _, item := range list {
		if item.Name == p.Name {
			p.projectid = item.ID
			p.rootcollid = item.RootCollectionId
		}
	}

	if p.rootcollid == "" && p.projectid == "" {
		return errors.New("invalid name")
	}

	p.token, err = GetProjectToken(p.projectid, p.rootcollid)
	if err != nil {
		return
	}

	p.root = &FileNode{
		Type:   Node_Dir,
		Name:   "/",
		NodeId: p.rootcollid,
		Child:  make(map[string]*FileNode),
	}
	p.collections(p.root.NodeId, p.root.Child, nil)
	return nil
}

func (p *ProjectFs) fixpath(path string) string {
	path = strings.TrimSpace(path)
	if path[0] != '/' {
		return ""
	}

	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	if strings.HasPrefix(path, "/"+p.Name) {
		path = path[len(p.Name)+1:]
	}

	return path
}

func (p *ProjectFs) collections(rootid string, root map[string]*FileNode, tokens []string, private ...interface{}) {
	colls, err := Collections(rootid, p.projectid)
	if err == nil {
		for _, coll := range colls {
			node := &FileNode{
				Type:     Node_Dir,
				Name:     coll.Title,
				NodeId:   coll.ID,
				ParentId: coll.ParentId,
				Updated:  coll.Updated,
				Child:    make(map[string]*FileNode),
				Private:  private,
			}
			if len(tokens) > 0 && node.Name == tokens[0] {
				p.collections(node.NodeId, node.Child, tokens[1:], private)
			}
			root[coll.Title] = node
		}
	}
}

func (p *ProjectFs) works(rootid string, root map[string]*FileNode, private ...interface{}) {
	works, err := Works(rootid, p.projectid)
	if err == nil {
		for _, work := range works {
			root[work.FileName] = &FileNode{
				Type:     Node_File,
				Name:     work.FileName,
				NodeId:   work.ID,
				ParentId: work.ParentId,
				Url:      work.DownloadUrl,
				Size:     work.FileSize,
				Updated:  work.Updated,
				Private:  private,
			}
		}
	}
}

func (p *ProjectFs) find(path string) (node *FileNode, prefix string, exist bool, err error) {
	newpath := p.fixpath(path)
	if newpath == "" {
		return node, prefix, exist, errors.New("invalid path")
	}

	defer func() {
		if err == nil && node != nil && node.Child == nil {
			node.Child = make(map[string]*FileNode)
		}
	}()

	tokens := strings.Split(newpath[1:], "/")
	node = p.root
	root := p.root.Child

	for idx, token := range tokens {
		if val, ok := root[token]; ok {
			node = val
			root = val.Child
			if idx == len(tokens)-1 {
				exist = true
				return node, "/" + strings.Join(tokens, "/"), exist, nil
			}
			continue
		}

		exist = false
		if idx == 0 {
			return node, "", exist, nil
		}

		return node, "/" + strings.Join(tokens[:idx], "/"), exist, nil
	}

	return
}

func (p *ProjectFs) mkdir(path string) (node *FileNode, err error) {
	newpath := p.fixpath(path)
	if newpath == "" {
		return node, errors.New("invalid path")
	}

	// query path
	accnode, prefix, exist, err := p.find(newpath)
	if err != nil {
		return node, err
	}
	if exist {
		p.works(accnode.NodeId, accnode.Child)
		return accnode, nil
	}

	p.mux.Lock()
	defer p.mux.Unlock()

	// sync accnode
	tokens := strings.Split(newpath[len(prefix)+1:], "/")
	p.collections(accnode.NodeId, accnode.Child, tokens)

	// query again
	accnode, prefix, exist, err = p.find(newpath)
	if err != nil {
		return node, err
	}
	if exist {
		p.works(accnode.NodeId, accnode.Child)
		return accnode, nil
	}

	// new path
	root := accnode
	tokens = strings.Split(newpath[len(prefix)+1:], "/")
	for _, token := range tokens {
		err = CreateCollection(root.NodeId, p.projectid, token)
		if err != nil {
			return node, err
		}
		list, err := Collections(root.NodeId, p.projectid)
		if err != nil {
			return node, err
		}

		if root.Child == nil {
			root.Child = make(map[string]*FileNode)
		}
		for _, coll := range list {
			if coll.Title == token {
				root.Child[coll.Title] = &FileNode{
					Type:     Node_Dir,
					Name:     token,
					ParentId: coll.ParentId,
					NodeId:   coll.ID,
					Updated:  coll.Updated,
				}
				root = root.Child[coll.Title]
				break
			}
		}
	}

	return root, nil
}

func (p *ProjectFs) UploadFile(filepath, targetdir string) error {
	if targetdir[0] != '/' {
		return errors.New("invalid dst")
	}

	info, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	node, err := p.mkdir(targetdir)
	if err != nil {
		return err
	}

	fmt.Println("node", node)

	filenode, exist := node.Child[info.Name()]
	if exist && filenode.Size == int(info.Size()) {
		return nil
	}
	if exist {
		err = Archive(filenode.NodeId)
		if err != nil {
			return err
		}
	}

	upload, err := UploadProjectFile(p.token, filepath)
	if err != nil {
		return err
	}

	return CreateWork(node.NodeId, upload)
}
func (p *ProjectFs) UploadDir(srcdir, targetdir string) error {
	_, err := os.Stat(srcdir)
	if err != nil {
		return err
	}

	srcdir, _ = filepath.Abs(srcdir)
	dirpaths := make(map[string][]string)
	curdir := ""
	filepath.Walk(srcdir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			curdir = path
			return nil
		}

		dirpaths[curdir] = append(dirpaths[curdir], path)
		return nil
	})

	wg := sync.WaitGroup{}
	count := 0
	for dir, files := range dirpaths {
		target := filepath.Join(targetdir, dir[len(srcdir):])
		for _, file := range files {
			count++
			wg.Add(1)
			f := file
			go func() {
				defer wg.Done()
				p.UploadFile(f, target)
			}()
			if count == 5 {
				wg.Wait()
				count = 0
			}
		}
	}

	if count > 0 {
		wg.Wait()
	}

	return nil
}
func (p *ProjectFs) DownloadUrl(srcpath, targetdir string) (error) {
	newpath := p.fixpath(srcpath)
	if newpath == "" {
		return errors.New("invalid path")
	}

	var tokens []string

	// query first
	accnode, prefix, exist, err := p.find(newpath)
	if err != nil {
		return err
	}
	if exist {
		goto download
	}

	// sync dirs
	p.mux.Lock()
	tokens = strings.Split(newpath[len(prefix)+1:], "/")
	p.collections(accnode.NodeId, accnode.Child, tokens)
	p.mux.Unlock()

	// query second
	accnode, prefix, exist, err = p.find(newpath)
	if err != nil {
		return err
	}
	if exist {
		goto download
	}

	// sync files
	p.mux.Lock()
	p.works(accnode.NodeId, accnode.Child)
	p.mux.Unlock()

	// query again
	accnode, prefix, exist, err = p.find(newpath)
	if err != nil || !exist {
		if err == nil {
			err = errors.New("not exist path")
		}
		return err
	}
	goto download

download:
	if accnode.Type == Node_File {
		return File(accnode.Url, "GET", nil, nil, filepath.Join(targetdir, accnode.Name))
	}
	return ArchiveProjectDir(p.token, accnode.NodeId, p.projectid, accnode.Name, targetdir)
}

func FindPanDir(dir string) (node Node, err error) {
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

func AutoLogin() {
	u, _ := url.Parse(www + "/")
	var session *http.Cookie
	for _, c := range jar.Cookies(u) {
		if c != nil && c.Name == "TEAMBITION_SESSIONID" {
			session = c
			break
		}
	}

	if session != nil {
		return
	}

	clientid, token, pubkey, err := LoginParams()
	if err != nil {
		return
	}

	remail := regexp.MustCompile(`^[A-Za-z0-9]+([_\\.][A-Za-z0-9]+)*@([A-Za-z0-9\-]+\.)+[A-Za-z]{2,6}$`)
	rphone := regexp.MustCompile(`^1[3-9]\\d{9}$`)
retry:
	var username string
	var password string
	fmt.Printf("Input Email/Phone:")
	fmt.Scanf("%s", &username)
	fmt.Printf("Input Password:")
	fmt.Scanf("%s", &password)

	if username == "" || password == "" {
		goto retry
	}

	if remail.MatchString(username) {
		_, err = Login(clientid, pubkey, token, username, "", password)
	} else if rphone.MatchString(username) {
		_, err = Login(clientid, pubkey, token, "", username, password)
	} else {
		goto retry
	}

	if err != nil {
		fmt.Println(err.Error())
		goto retry
	}

	fmt.Println("登录成功!!!")
}

func main() {
	AutoLogin()
	UserAgent = agents[0]

	_, _, _, err := GetCacheData()
	if err != nil {
		log.Println(err)
		return
	}

	p := ProjectFs{Name: "data", Orgid: "5f6707e0f0aab521364694ee"}
	fmt.Println(p.Init())
	fmt.Println(p.UploadDir("/home/user/go/src/golearn/http","/wx"))

	//dir := "/packages"
	//log.Println("Making Dir", dir)
	//nodeid, err := PanMkdirP(dir)
	//if err != nil {
	//	log.Println("PanMkdirP", err)
	//	return
	//}
	//log.Println("Success")
	//
	//var files = []string{"/home/user/Downloads/gz/PgyVPN_Ubuntu_2.2.1_X86_64.deb"}
	//for _, file := range files {
	//	log.Println("Starting CreateFile:", file)
	//	files, err := CreatePanFile(org.OrganizationId, nodeid, spaces[0].SpaceId, org.DriveId, file)
	//	if err == nil {
	//		log.Println("Starting UploadUrl ...")
	//		upload, err := CreatePanUpload(org.OrganizationId, files[0])
	//		if err == nil {
	//			log.Println("Starting UploadPanFile ...")
	//			UploadPanFile(org.OrganizationId, upload, file)
	//			log.Println("Success")
	//		}
	//	}
	//
	//	time.Sleep(500 * time.Millisecond)
	//}
	//
	//time.Sleep(5 * time.Second)
}
