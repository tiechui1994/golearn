package compress

import (
	"compress/zlib"
	"bytes"
	"io"
	"os"
	"archive/zip"
	"path"
	"fmt"
	"errors"
)

func Zip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw, err := zlib.NewWriterLevel(&buf, zlib.DefaultCompression)
	if err != nil {
		return nil, err
	}
	_, err = zw.Write(data)
	if err != nil {
		return nil, err
	}

	err = zw.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func UnZip(data []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zr)
	return buf.Bytes(), err
}

func ZipFile(dest string, src []string) error {
	writer, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer writer.Close()

	zw := zip.NewWriter(writer)

	var buf = make([]byte, 1024)
	for _, file := range src {
		reader, err := os.Open(file)
		if err != nil {
			continue
		}
		_, name := path.Split(reader.Name())
		writer, err := zw.Create(name)
		if err != nil {
			reader.Close()
			continue
		}

		_, err = io.CopyBuffer(writer, reader, buf)
		if err != nil {
			reader.Close()
			continue
		}

		reader.Close()
	}

	return zw.Close()
}

func UnZipFile(destdir string, src string) error {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	stats, _ := reader.Stat()
	zr, err := zip.NewReader(reader, stats.Size())
	if err != nil {
		return err
	}

	var msg bytes.Buffer
	var buf = make([]byte, 1024)
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			dest := path.Join(destdir, file.Name)
			os.Mkdir(dest, 0777)
			continue
		}

		reader, err := file.Open()
		if err != nil {
			fmt.Fprintf(&msg, "[Open] %v \n", err)
			continue
		}

		dest := path.Join(destdir, file.Name)
		writer, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			reader.Close()
			fmt.Fprintf(&msg, "[OpenFile] %v \n", err)
			continue
		}

		_, err = io.CopyBuffer(writer, reader, buf)
		if err != nil {
			fmt.Fprintf(&msg, "[CopyBuffer] %v \n", err)
			reader.Close()
			writer.Close()
			continue
		}

		reader.Close()
		writer.Close()
	}

	if msg.Len() > 0 {
		return errors.New(msg.String())
	}

	return nil
}
