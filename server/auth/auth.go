package auth

import (
	"net/http"
)

type Authenticator interface {
	IsAuthorized(r *http.Request) bool
}

// BasicAuth uses HTTP basic auth to authenticate and authorize a user.
type BasicAuth struct {
}

func NewBasicAuth() *BasicAuth {
	return &BasicAuth{}
}

func (b *BasicAuth) IsAuthorized(r *http.Request) bool {
	return false
}
