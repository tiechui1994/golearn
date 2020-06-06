# go 压缩算法的使用

- zlib

`compress/zlib`, 实现了 `zip` 压缩算法. (无损压缩算法)

文本压缩:

```cgo
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
```

> level参数值(常用的值):
>
> - zlib.NoCompression (0), 不压缩
> - zlib.BestSpeed (1), 最快的压缩速度
> - zlib.BestCompression (9), 最好的压缩级别
> - zlib.DefaultCompression (-1), 使用默认的压缩级别 (6). 
> - zlib.HuffmanOnly (-2), 使用 `哈夫曼压缩算法`


文件压缩(可以压缩多个文件):

> 使用的是 `archive/zip` 包

```cgo
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
```

---

- gzip 

`compress/gzip`, 实现了 `gzip` 压缩算法. (无损压缩算法)

文本压缩:

```cgo
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
```

> level参数值(常用的值), 和 `zip` 的类似.


文件压缩(只能压缩一个文件):

```cgo
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
```