package settings

import (
	"errors"
	"net/http"

	"appengine"
)

var ErrUnsupportedType = errors.New("unsupported type")

type SettingsService struct {
}

func NewSettingsService() *SettingsService {
	return &SettingsService{}
}

type SetAuthRequest struct {
	Type string `json:"type"`
}

type SetAuthResponse struct {
	Success bool `json:"success"`
}

func (s *SettingsService) SetAuth(r *http.Request, request *SetAuthRequest, response *SetAuthResponse) error {
	c := appengine.NewContext(r)

	var err error
	if request.Type == "basic" {
		values := make(Settings)
		values["type"] = "basic"
		err = Set(c, "auth", values)
	} else {
		err = ErrUnsupportedType
	}

	response.Success = err == nil
	return err
}

type SetStorageRequest struct {
	Type   string `json:"type"`
	Bucket string `json:"bucket"`
}

type SetStorageResponse struct {
	Success bool `json:"success"`
}

func (s *SettingsService) SetStorage(r *http.Request, request *SetStorageRequest, response *SetStorageResponse) error {
	c := appengine.NewContext(r)

	var err error
	if request.Type == "gcs" {
		values := make(Settings)
		values["type"] = "gcs"
		values["bucket"] = request.Bucket
		err = Set(c, "storage", values)
	} else {
		err = ErrUnsupportedType
	}

	response.Success = err == nil
	return err
}
