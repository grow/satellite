package server

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"path"

	"server/auth"
	"server/storage"
)

type SatelliteServer struct {
	auth  auth.Authenticator
	files storage.FileStorage
}

func (s *SatelliteServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Prevent click-jacking.
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")

	// Authorize the request.
	is_authorized := s.auth.IsAuthorized(r)
	if !is_authorized {
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
	if !s.files.Exists(filePath) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "404: Not Found")
		return
	}

	// Set the Content-Type header based on the file ext.
	mimetype := mime.TypeByExtension(ext)
	log.Println(mimetype)
	if mimetype != "" {
		w.Header().Set("Content-Type", mimetype)
	}

	// Write the contents of the file to the response.
	s.files.Read(filePath, w)
}

func init() {
	mime.AddExtensionType(".ico", "image/x-icon")

	// TODO(stevenle): Read configuration settings from datastore.
	var authenticator auth.Authenticator
	var files storage.FileStorage
	authenticator = auth.NewBasicAuth()
	files = storage.NewGcsFileStorage("bucket")

	s := &SatelliteServer{
		auth:  authenticator,
		files: files,
	}

	// TODO(stevenle): Add an admin handler for configuration changes.
	http.Handle("/", s)
}
