package services

import (
	"net/http"

	"appengine"
	"server/auth"
	"server/domains"
)

type BasicAuthService struct {
}

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
	d, err := domains.Get(r)
	if err != nil {
		c := appengine.NewContext(r)
		c.Errorf("domain error: %v", err)
		response.Success = false
		return err
	}

	c := d.Context()
	a := auth.NewBasicAuth(c)
	err = a.AddUser(request.Username, request.Password)
	if err != nil {
		c.Errorf("add user error: %v", err)
		response.Success = false
		return err
	}

	response.Success = true
	return nil
}
