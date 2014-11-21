package auth

import (
	"net/http"

	"appengine"
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
	c := appengine.NewContext(r)
	a := NewBasicAuth()
	err := a.AddUser(c, request.Username, request.Password)
	response.Success = err == nil
	return err
}
