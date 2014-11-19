package server

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
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

	// Force https.
	if r.URL.Scheme != "https" && !appengine.IsDevAppServer() {
		redirectUrl, _ := url.ParseRequestURI(r.URL.String())
		redirectUrl.Scheme = "https"
		http.Redirect(w, r, redirectUrl.String(), http.StatusMovedPermanently)
		return
	}

	// Initialize the server.
	if !s.initialized {
		settings := make([]Settings, 2)
		GetSettings(c, []string{"auth", "storage"}, settings)
		authSettings := settings[0]
		storageSettings := settings[1]

		if authSettings == nil || storageSettings == nil {
			// Redirect user to the admin configuration page.
			http.Redirect(w, r, "/admin/settings", http.StatusFound)
			return
		}

		if authSettings["type"] == "basic" {
			basicAuth := auth.NewBasicAuth()
			if appengine.IsDevAppServer() {
				go basicAuth.AddUser(c, "test", "testing123")
			}
			s.auth = basicAuth
		} else {
			c.Errorf("Unknown auth type: %v", authSettings["type"])
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "500: Internal Server Error")
			return
		}

		if storageSettings["type"] == "gcs" {
			s.files = storage.NewGcsFileStorage(storageSettings["bucket"])
		} else {
			c.Errorf("Unknown storage type: %v", storageSettings["type"])
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "500: Internal Server Error")
			return
		}

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
