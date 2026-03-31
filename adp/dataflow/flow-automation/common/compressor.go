package common

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// File 待压缩文件信息
type File struct {
	Reader   io.Reader
	FileName string
	FileSize int64
}

// Compress 压缩文件
func Compress(buf *bytes.Buffer, files []File) error {
	// 创建 gzip writer，将其链接到 tar writer
	gzipWriter := gzip.NewWriter(buf)
	defer gzipWriter.Close()

	// 创建 tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, file := range files {
		// 在压缩文件中创建一个文件
		header := &tar.Header{
			Name: filepath.Base(file.FileName),
			Mode: 0666,
			Size: file.FileSize,
		}

		// 写入文件头
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		_, err := io.Copy(tarWriter, file.Reader)
		if err != nil {
			return err
		}
	}

	return nil
}

// Decompress 解压文件
func Decompress(src, dest string) error {
	if !strings.HasSuffix(dest, "/") {
		dest = fmt.Sprintf("%s/", dest)
	}
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	// 创建 gzip reader
	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// 创建 tar writer
	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := dest + header.Name

		// 如果是目录，则创建目录
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(path, header.FileInfo().Mode()); err != nil {
				return err
			}
			continue
		}

		// 如果是文件，则读取内容并写入到对应路径的文件中
		fileToWrite, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, header.FileInfo().Mode())
		if err != nil {
			return err
		}
		defer fileToWrite.Close()

		_, err = io.Copy(fileToWrite, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}
