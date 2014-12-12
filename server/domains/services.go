package domains

import (
	"net/http"

	"appengine"
	"appengine/datastore"
)

type DomainService struct {
}

func NewDomainService() *DomainService {
	return &DomainService{}
}

type SetDomainRequest struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

type SetDomainResponse struct {
	Success bool `json:"success"`
}

func (s *DomainService) SetDomain(r *http.Request, request *SetDomainRequest, response *SetDomainResponse) error {
	c := appengine.NewContext(r)
	k := datastore.NewKey(c, DomainEntityType, request.Name, 0 /* intID */, nil /* parent */)
	e := &DomainEntity{
		Name:    request.Name,
		Aliases: request.Aliases,
	}
	_, err := datastore.Put(c, k, e)
	if err != nil {
		c.Errorf("set domain error: %v", err)
	}
	response.Success = err == nil
	return err
}
