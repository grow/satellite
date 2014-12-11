package storage

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"appengine"
	"appengine/blobstore"
	gcs "appengine/file"
	"appengine/memcache"
	"appengine/urlfetch"
	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/storage/v1"
)

type FileStorage interface {
	Serve(w http.ResponseWriter, r *http.Request) error
}

type GcsFileStorage struct {
	bucket string
}

type GcsFileStat struct {
	Etag         string    `json:"etag"`
	ModifiedTime time.Time `json:"modified"`
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

func (g *GcsFileStorage) Serve(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)

	filePath := r.URL.Path
	ext := path.Ext(filePath)
	if ext == "" {
		filePath = path.Join(filePath, "index.html")
		ext = ".html"
	}
	gcsPath := g.getGcsPath(filePath)

	// Get file stat.
	stat, err := g.Stat(c, filePath)
	if err != nil {
		c.Errorf("stat error: %v", err)
	}
	if stat == nil {
		w.WriteHeader(http.StatusNotFound)
		// TODO(stevenle): add custom 404 pages.
		fmt.Fprintln(w, "404: Not Found")
		return nil
	}

	// Set the Content-Type header based on the file ext.
	mimetype := mime.TypeByExtension(ext)
	if mimetype != "" {
		w.Header().Set("Content-Type", mimetype)
	}

	// By default, set cache-control headers for images.
	if strings.HasPrefix(mimetype, "image/") {
		w.Header().Set("Cache-Control", "private, max-age=3600, s-maxage=3600")
	}

	// Set HTTP modification time and etag headers.
	w.Header().Set("ETag", stat.Etag)
	w.Header().Set("Last-Modified", stat.ModifiedTime.Format(time.RFC1123))

	if r.Header.Get("If-None-Match") == stat.Etag {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	blobKey, err := blobstore.BlobKeyForFile(c, gcsPath)
	if err != nil {
		return err
	}
	blobstore.Send(w, blobKey)
	return nil
}

func (g *GcsFileStorage) Stat(c appengine.Context, filePath string) (*GcsFileStat, error) {
	// Check for cached value.
	cacheKey := "stat:" + filePath
	item, err := memcache.Get(c, cacheKey)
	if err == nil {
		var stat GcsFileStat
		err = json.Unmarshal(item.Value, &stat)
		if err == nil {
			return &stat, nil
		}
	}

	// Remove leading slash to ge the GCS object id.
	// E.g. filePath = "/index.html", gcsObjectId = "index.html".
	gcsObjectId := filePath[1:]

	accessToken, _, err := appengine.AccessToken(c, storage.CloudPlatformScope, storage.DevstorageRead_onlyScope)
	if err != nil {
		return nil, err
	}

	// TODO(stevenle): concurrently read from memcache.
	transport := &oauth.Transport{
		Token: &oauth.Token{
			AccessToken: accessToken,
		},
		Transport: &urlfetch.Transport{
			Context: c,
		},
	}
	client := &http.Client{Transport: transport}
	storageService, err := storage.New(client)
	if err != nil {
		return nil, err
	}

	objectService := storage.NewObjectsService(storageService)
	obj, err := objectService.Get(g.bucket, gcsObjectId).Do()
	if err != nil {
		return nil, err
	}

	modTime, err := time.Parse(time.RFC3339, obj.Updated)
	if err != nil {
		return nil, err
	}

	stat := &GcsFileStat{
		Etag:         obj.Etag,
		ModifiedTime: modTime,
	}

	// Set value to cache.
	done := make(chan bool)
	go func() {
		cacheValue, err := json.Marshal(stat)
		if err == nil {
			cacheItem := &memcache.Item{
				Key:   cacheKey,
				Value: cacheValue,
			}
			memcache.Set(c, cacheItem)
		}
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Millisecond):
		c.Errorf("memcache timeout")
	}

	return stat, nil
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
