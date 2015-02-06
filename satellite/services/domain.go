package services

import (
	"net/http"

	"appengine"
	"satellite/domains"
)

type DomainService struct{}

func NewDomainService() *DomainService {
	return &DomainService{}
}

type SetDomainRequest struct {
	Domain domains.DomainEntity `json:"domain"`
}

type SetDomainResponse struct {
	Success bool `json:"success"`
}

func (s *DomainService) SetDomain(r *http.Request, request *SetDomainRequest, response *SetDomainResponse) error {
	c := appengine.NewContext(r)
	err := domains.Put(c, &request.Domain)
	if err != nil {
		c.Errorf("set domain error: %v", err)
	}
	response.Success = err == nil
	return err
}
