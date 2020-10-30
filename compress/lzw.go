package compress

import (
	"bytes"
	"compress/lzw"
	"io"
	"os"
)

func Lzw(data []byte) ([]byte, error) {
	var writer bytes.Buffer
	lw := lzw.NewWriter(&writer, lzw.LSB, 8)

	_, err := lw.Write(data)
	if err != nil {
		return nil, err
	}

	err = lw.Close()
	if err != nil {
		return nil, err
	}

	return writer.Bytes(), nil
}

func LzwFile(dest string, src string) error {
	writer, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer writer.Close()

	lw := lzw.NewWriter(writer, lzw.LSB, 8)

	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(lw, reader)
	if err != nil {
		return err
	}

	return lw.Close()
}

func UnLzw(data []byte) ([]byte, error) {
	lr := lzw.NewReader(bytes.NewReader(data), lzw.LSB, 8)
	defer lr.Close()

	var writer bytes.Buffer
	_, err := io.Copy(&writer, lr)
	return writer.Bytes(), err
}

func UnLzwFile(dest string, src string) error {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	lr := lzw.NewReader(reader, lzw.LSB, 8)
	defer lr.Close()

	writer, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, lr)
	return err
}
