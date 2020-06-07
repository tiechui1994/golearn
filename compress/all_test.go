package compress

import (
	"testing"
	"bytes"
	"encoding/hex"
)

func TestZip(t *testing.T) {
	origin := []byte("Hello World")
	zipdata, err := Zip(origin)
	if err != nil {
		t.Fatalf("Zip:%v", err)
	}

	uzipdata, err := UnZip(zipdata)
	if err != nil {
		t.Fatalf("UnZip:%v", err)
	}

	t.Logf("Equal: %v", bytes.Equal(origin, uzipdata))
}

func TestGzip(t *testing.T) {
	origin := []byte("Hello")
	gzipdata, err := Gzip(origin)
	if err != nil {
		t.Fatalf("Gzip:%v", err)
	}

	t.Logf("%v", hex.EncodeToString(gzipdata))

	ugzipdata, err := UnGzip(gzipdata)
	if err != nil {
		t.Fatalf("UnGzip:%v", err)
	}

	t.Logf("Equal: %v", bytes.Equal(origin, ugzipdata))
	t.Logf("origin: %v", hex.EncodeToString(origin))
	t.Logf("ungzip: %v", hex.EncodeToString(ugzipdata))
}

func TestGzipFile(t *testing.T) {
	err := GzipFile("/tmp/xx.gz",
		"/home/quinn/go/src/golearn/compress/gzip.go")
	if err != nil {
		t.Fatalf("GzipFile: %v", err)
	}
}

func TestUnGzipFile(t *testing.T) {
	err := UnGzipFile("/tmp", "/home/quinn/Desktop/zip.go.gz")
	if err != nil {
		t.Fatalf("UnGzipFile: %v", err)
	}
}

func TestZipFile(t *testing.T) {
	err := ZipFile("/tmp/xx.zip", []string{
		"/home/quinn/go/src/golearn/compress/gzip.go",
		"/home/quinn/go/src/golearn/compress/zip.go",
	})
	if err != nil {
		t.Fatalf("ZipFile: %v", err)
	}
}

func TestUnZipFile(t *testing.T) {
	err := UnZipFile("/tmp", "/home/quinn/Desktop/t.zip")
	if err != nil {
		t.Fatalf("[UnZipFile]\n%v", err)
	}
}

func TestUnBizp2File(t *testing.T) {
	err := UnBizp2File("/tmp/xx.go", "/home/quinn/Desktop/w/zip.go.bz2")
	if err != nil {
		t.Fatalf("[UnBizp2File] %v", err)
	}
}

func TestLzw(t *testing.T) {
	data := []byte("Hello")

	lzwdata, err := Lzw(data)
	if err != nil {
		t.Fatalf("Lzw %v", err)
	}
	t.Logf("data len: %v", len(data))
	t.Logf("lzw  len: %v", len(lzwdata))

	lzrdata, err := UnLzw(lzwdata)
	if err != nil {
		t.Fatalf("UnLzw %v", err)
	}

	t.Logf("Equal: %v", bytes.Equal(lzrdata, data))
}

func TestLzwFile(t *testing.T) {
	err := LzwFile("/tmp/xx.lz", "/home/quinn/go/src/golearn/compress/lzw.go")
	if err != nil {
		t.Fatalf("LzwFile %v", err)
	}

	err = UnLzwFile("/tmp/xx.go", "/tmp/xx.lz")
	if err != nil {
		t.Fatalf("UnLzwFile %v", err)
	}
}

func TestTgz(t *testing.T) {
	err := Tgz("/tmp/xxx.tar.gz", []string{
		"/home/quinn/go/src/golearn/compress/bzip2.go",
		"/home/quinn/go/src/golearn/compress/lzw.go",
	})
	if err != nil {
		t.Fatalf("Tgz %v", err)
	}

	err = UnTgz("/tmp/", "/tmp/x.tar.gz")
	if err != nil {
		t.Fatalf("UnTgz %v", err)
	}
}
