package server

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"

	"appengine"
	"github.com/gorilla/rpc/v2"
	jsonrpc "github.com/gorilla/rpc/v2/json"
	"server/auth"
	"server/settings"
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
		settingsList := make([]settings.Settings, 2)
		settings.GetMulti(c, []string{"auth", "storage"}, settingsList)
		authSettings := settingsList[0]
		storageSettings := settingsList[1]

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
		// Try to render /404.html if it exists.
		if s.files.Exists(c, "/404.html") {
			s.files.Serve(c, "/404.html", w)
		} else {
			fmt.Fprintln(w, "404: Not Found")
		}
		return
	}

	// Set the Content-Type header based on the file ext.
	mimetype := mime.TypeByExtension(ext)
	if mimetype != "" {
		w.Header().Set("Content-Type", mimetype)
	}

	// By default, set cache-control headers for images.
	if strings.HasPrefix(mimetype, "image/") {
		w.Header().Set("Cache-Control", "private, max-age=3600")
	}

	// Serve the file.
	err := s.files.Serve(c, filePath, w)
	if err != nil {
		// Internal server error.
		// Try to render /500.html if it exists.
		c.Errorf("Failed to serve %v: %v", filePath, err)
		w.WriteHeader(http.StatusInternalServerError)
		if s.files.Exists(c, "/500.html") {
			s.files.Serve(c, "/500.html", w)
		} else {
			fmt.Fprintln(w, "500: Internal Server Error")
		}
		return
	}
}

func init() {
	mime.AddExtensionType(".ico", "image/x-icon")
	s := &SatelliteServer{
		initialized: false,
	}

	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	rpcServer.RegisterService(auth.NewBasicAuthService(), "BasicAuthService")
	rpcServer.RegisterService(settings.NewSettingsService(), "SettingsService")

	http.Handle("/admin/rpc", rpcServer)
	http.Handle("/", s)
}
