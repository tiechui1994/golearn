package compress

import (
	"compress/bzip2"
	"io"
	"os"
)

func UnBizp2File(dest string, src string) (error) {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	br := bzip2.NewReader(reader)

	writer, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	_, err = io.Copy(writer, br)
	return err
}
