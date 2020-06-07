package compress

import (
	"archive/tar"
	"os"
	"io"
	"bytes"
	"compress/gzip"
	"path"
	"fmt"
	"errors"
)

func Tgz(dest string, src []string) error {
	// tar
	var wr bytes.Buffer
	tw := tar.NewWriter(&wr)

	buf := make([]byte, 1024)
	for _, file := range src {
		reader, err := os.Open(file)
		if err != nil {
			continue
		}

		header := new(tar.Header)
		info, _ := reader.Stat()
		header.Size = info.Size()
		header.Name = info.Name()
		err = tw.WriteHeader(header)
		if err != nil {
			reader.Close()
			continue
		}
		io.CopyBuffer(tw, reader, buf)
		reader.Close()
	}

	err := tw.Close()
	if err != nil {
		return err
	}

	// gzip
	writer, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer writer.Close()

	gw, err := gzip.NewWriterLevel(writer, gzip.BestCompression)
	if err != nil {
		return err
	}

	_, err = io.CopyBuffer(gw, &wr, buf)
	if err != nil {
		return err
	}

	return gw.Close()
}

func UnTgz(destdir string, src string) error {
	// gz
	reader, err := os.Open(src)
	if err != nil {
		return err
	}

	gr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gr.Close()

	// tar
	tw := tar.NewReader(gr)

	var msg bytes.Buffer
	header, err := tw.Next()
	if err != nil {
		return err
	}

	for header != nil {
		info := header.FileInfo()
		if info.IsDir() {
			err = os.Mkdir(path.Join(destdir, info.Name()), 0775)
			if err != nil {
				fmt.Fprintf(&msg, "[Mkdir] %v \n", err)
			}
			header, err = tw.Next()
			continue
		}

		dest := path.Join(destdir, info.Name())
		writer, err := os.Create(dest)
		if err != nil {
			fmt.Fprintf(&msg, "[Create] %v \n", err)
			header, err = tw.Next()
			continue
		}

		_, err = io.CopyN(writer, tw, info.Size())
		if err != nil {
			fmt.Fprintf(&msg, "[CopyN] %v \n", err)
		}
		writer.Close()
		header, err = tw.Next()
	}

	if msg.Len() > 0 {
		return errors.New(msg.String())
	}

	return nil
}
