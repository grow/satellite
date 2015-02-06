package satellite

import (
	"fmt"
	"mime"
	"net/http"

	"appengine"
	"github.com/gorilla/rpc/v2"
	jsonrpc "github.com/gorilla/rpc/v2/json"
	"satellite/domains"
	"satellite/services"
)

type SatelliteServer struct{}

func (s *SatelliteServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	d, err := domains.Get(c, r.URL.Host)
	if err != nil {
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

	r := rpc.NewServer()
	r.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	r.RegisterService(services.NewBasicAuthService(), "BasicAuthService")
	r.RegisterService(services.NewDomainService(), "DomainService")
	http.Handle("/_/rpc", r)

	s := &SatelliteServer{}
	http.Handle("/", s)
}
