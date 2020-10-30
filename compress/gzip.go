package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path"
)

func Gzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}

	_, err = gw.Write(data)
	if err != nil {
		return nil, err
	}

	// todo: 必须调用 Close(), 这样才能写入所有的所有的数据
	err = gw.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func UnGzip(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gr)
	return buf.Bytes(), err
}

func GzipFile(dest string, src string) error {
	writer, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer writer.Close()

	gw, err := gzip.NewWriterLevel(writer, gzip.BestCompression)
	if err != nil {
		return err
	}

	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, gw.Name = path.Split(reader.Name())
	var buf = make([]byte, 1024)
	_, err = io.CopyBuffer(gw, reader, buf)
	if err != nil {
		return err
	}

	return gw.Close()
}

func UnGzipFile(destdir string, src string) error {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	gr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}

	dest := path.Join(destdir, gr.Name)
	writer, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer writer.Close()

	var buf = make([]byte, 1024)
	_, err = io.CopyBuffer(writer, gr, buf)
	if err != nil {
		return err
	}

	return gr.Close()
}
