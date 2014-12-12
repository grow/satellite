package services

import (
	"net/http"

	"appengine"
	"satellite/auth"
	"satellite/domains"
)

type BasicAuthService struct{}

func NewBasicAuthService() *BasicAuthService {
	return &BasicAuthService{}
}

type AddUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AddUserResponse struct {
	Success bool `json:"success"`
}

func (b *BasicAuthService) AddUser(r *http.Request, request *AddUserRequest, response *AddUserResponse) error {
	c := appengine.NewContext(r)
	d, err := domains.Get(c, r.URL.Host)
	if err != nil {
		c.Errorf("domain error: %v", err)
		response.Success = false
		return err
	}

	a := auth.NewBasicAuth(d.Context())
	err = a.AddUser(request.Username, request.Password)
	if err != nil {
		c.Errorf("add user error: %v", err)
		response.Success = false
		return err
	}

	response.Success = true
	return nil
}
