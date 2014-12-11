package storage

import (
	"testing"

	"appengine"
	"appengine/aetest"
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
}
