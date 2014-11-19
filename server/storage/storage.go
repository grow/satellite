package storage

import (
	"bufio"
	"io"
)

type FileStorage interface {
	Exists(path string) bool
	Read(path string, w io.Writer)
}

type GcsFileStorage struct {
	bucket string
}

func NewGcsFileStorage(bucket string) *GcsFileStorage {
	return &GcsFileStorage{
		bucket: bucket,
	}
}

func (g *GcsFileStorage) Exists(path string) bool {
	return true
}

func (g *GcsFileStorage) Read(path string, w io.Writer) {
	buffer := bufio.NewWriter(w)
	buffer.Write([]byte("Hello, world!"))
	buffer.Flush()
}
