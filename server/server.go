package server

import (
	"fmt"
	"net/http"

	"server/auth"
)

type SatelliteServer struct {
	authenticator auth.Authenticator
}

func (s *SatelliteServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	is_authorized := s.authenticator.IsAuthorized(r)
	if !is_authorized {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "401: Unauthorized\n")
		return
	}

	// TODO(stevenle): Render static files from a GCS bucket.
}

func init() {
	// TODO(stevenle): Read configuration settings from datastore.
	var authenticator auth.Authenticator
	authenticator = auth.NewBasicAuth()

	s := &SatelliteServer{
		authenticator: authenticator,
	}
	http.Handle("/", s)
}
