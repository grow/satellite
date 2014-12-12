package server

import (
	"fmt"
	"mime"
	"net/http"

	"appengine"
	"github.com/gorilla/rpc/v2"
	jsonrpc "github.com/gorilla/rpc/v2/json"
	authservice "server/auth/services"
	"server/domains"
)

type SatelliteServer struct {
	initialized bool
}

func (s *SatelliteServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d, err := domains.Get(r)
	if err != nil {
		c := appengine.NewContext(r)
		c.Errorf("domain error: %v: %v", r.URL.Host, err)

		// TODO(stevenle): add a user-friendly admin ui for setup / configuration.
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "not configured", r.URL.Host)
		return
	}
	d.ServeHTTP(w, r)
}

func init() {
	mime.AddExtensionType(".ico", "image/x-icon")
	s := &SatelliteServer{
		initialized: false,
	}

	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	rpcServer.RegisterService(authservice.NewBasicAuthService(), "BasicAuthService")
	rpcServer.RegisterService(domains.NewDomainService(), "DomainService")

	http.Handle("/_/rpc", rpcServer)
	http.Handle("/", s)
}
