package storage

import (
	"mime"
	"net/http"
	"path"

	"appengine"
	"appengine/blobstore"
	gcs "appengine/file"
)

type FileStorage interface {
	Exists(c appengine.Context, filePath string) bool
	Serve(c appengine.Context, filePath string, w http.ResponseWriter) error
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

func (g *GcsFileStorage) Serve(c appengine.Context, filePath string, w http.ResponseWriter) error {
	gcsPath := g.getGcsPath(filePath)
	blobKey, err := blobstore.BlobKeyForFile(c, gcsPath)
	if err != nil {
		return err
	}
	blobstore.Send(w, blobKey)
	return nil
}

func (g *GcsFileStorage) Write(c appengine.Context, filePath string, content []byte) error {
	ext := path.Ext(filePath)
	mimetype := mime.TypeByExtension(ext)
	opts := &gcs.CreateOptions{
		MIMEType:   mimetype,
		BucketName: g.bucket,
	}

	gcsPath := g.getGcsPath(filePath)
	writer, _, err := gcs.Create(c, gcsPath, opts)
	if err != nil {
		return err
	}

	defer writer.Close()
	_, err = writer.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func (g *GcsFileStorage) getGcsPath(filePath string) string {
	return path.Join("/gs", g.bucket, filePath)
}
