package storage

import (
	"io"
	"path"

	"appengine"
	gcs "appengine/file"
)

type FileStorage interface {
	Exists(c appengine.Context, filePath string) bool
	Open(c appengine.Context, filePath string) (io.Reader, error)
}

type GcsFileStorage struct {
	bucket string
}

func NewGcsFileStorage(bucket string) *GcsFileStorage {
	return &GcsFileStorage{
		bucket: bucket,
	}
}

func (g *GcsFileStorage) Exists(c appengine.Context, filePath string) bool {
	gcsPath := g.getGcsPath(filePath)
	fileInfo, _ := gcs.Stat(c, gcsPath)
	return fileInfo != nil
}

func (g *GcsFileStorage) Open(c appengine.Context, filePath string) (io.Reader, error) {
	gcsPath := g.getGcsPath(filePath)
	return gcs.Open(c, gcsPath)
}

func (g *GcsFileStorage) getGcsPath(filePath string) string {
	return path.Join("/gs", g.bucket, filePath)
}
