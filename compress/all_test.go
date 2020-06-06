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
