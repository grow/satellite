package server

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"

	"appengine"
	"server/auth"
	"server/storage"
)

type SatelliteServer struct {
	initialized bool
	auth        auth.Authenticator
	files       storage.FileStorage
}

func (s *SatelliteServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// Prevent click-jacking.
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")

	if !s.initialized {
		// TODO(stevenle): Read configuration settings from datastore.
		// For now, use basic auth and GCS for auth and storage.
		s.auth = auth.NewBasicAuth()
		s.files = storage.NewGcsFileStorage("app_default_bucket") // devappserver
		s.initialized = true
	}

	// Authorize the request.
	authorized := s.auth.IsAuthorized(r)
	if !authorized {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"Please enter a username and password\"")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "401: Unauthorized")
		return
	}

	// Determine the file path from the URL.
	filePath := r.URL.Path
	ext := path.Ext(filePath)
	if ext == "" {
		filePath = path.Join(filePath, "index.html")
		ext = ".html"
	}
	if !s.files.Exists(c, filePath) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "404: Not Found")
		return
	}

	// Open the file from storage.
	reader, err := s.files.Open(c, filePath)
	if err != nil {
		c.Errorf("Failed to open %v: %v", filePath, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "500: Internal Server Error")
		return
	}

	// Set the Content-Type header based on the file ext.
	mimetype := mime.TypeByExtension(ext)
	if mimetype != "" {
		w.Header().Set("Content-Type", mimetype)
	}

	// Write the contents of the file to the response.
	io.Copy(w, reader)
}

func init() {
	mime.AddExtensionType(".ico", "image/x-icon")
	s := &SatelliteServer{
		initialized: false,
	}
	// TODO(stevenle): Add an admin handler for configuration changes.
	http.Handle("/", s)
}
