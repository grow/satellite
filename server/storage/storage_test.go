package storage

import (
	"net/http/httptest"
	"testing"

	"appengine"
	"appengine/aetest"
	"appengine/blobstore"
)

func TestGcsStorage(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	s := NewGcsFileStorage("bucket")
	if s == nil {
		t.Fatal("Failed to create GCS storage")
	}

	gcsPath := s.getGcsPath("/index.html")
	if gcsPath != "/gs/bucket/index.html" {
		t.Fatalf("Wrong GCS path: %s", gcsPath)
	}

	// Write /hello.txt.
	req1, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create req1: %v", err)
	}
	c1 := appengine.NewContext(req1)
	err = s.Write(c1, "/hello.txt", []byte("Hello, world!"))
	if err != nil {
		t.Fatalf("Failed to write /hello.txt: %v", err)
	}

	// Verify /hello.txt exists.
	req2, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create req2: %v", err)
	}
	c2 := appengine.NewContext(req2)
	if !s.Exists(c2, "/hello.txt") {
		t.Fatal("Failed: /hello.txt does not exist")
	}

	req3, err := inst.NewRequest("GET", "/hello.txt", nil)
	if err != nil {
		t.Fatal("Failed to create req3: %v", err)
	}
	c3 := appengine.NewContext(req3)
	blobKey, err := blobstore.BlobKeyForFile(c3, s.getGcsPath("/hello.txt"))
	if err != nil {
		t.Fatalf("Failed to get blob key for /hello.txt: %v", err)
	}
	w := httptest.NewRecorder()
	s.Serve(c3, "/hello.txt", w)
	if w.Header().Get("X-AppEngine-BlobKey") != string(blobKey) {
		t.Fatal("Failed to send response via blobstore")
	}
}
