package domains

import (
	"fmt"
	"net/http"

	"appengine"
	"appengine/datastore"
	"server/auth"
	"server/storage"
)

const DomainEntityType = "Domain"

type DomainEntity struct {
	Name            string          `datastore:"name"`
	Aliases         []string        `datastore:"aliases"`
	AuthSettings    AuthSettings    `datastore:"auth"`
	StorageSettings StorageSettings `datastore:"storage"`
}

type AuthSettings struct {
	Type string `datastore:"type"`
}

type StorageSettings struct {
	Type string `datastore:"type"`

	// GcsFileStorage settings.
	Bucket string `datastore:"bucket"`
}

type Domain struct {
	context appengine.Context
	name    string
	entity  *DomainEntity
}

func Get(r *http.Request) (*Domain, error) {
	c := appengine.NewContext(r)

	// TODO(stevenle): support domain aliases.
	name := r.URL.Host
	k := datastore.NewKey(c, DomainEntityType, name, 0 /* intID */, nil /* parent */)
	e := new(DomainEntity)
	err := datastore.Get(c, k, e)
	if err != nil {
		return nil, err
	}

	d := &Domain{
		context: c,
		name:    name,
		entity:  e,
	}
	return d, nil
}

func (d *Domain) Auth() auth.Authenticator {
	if d.entity.AuthSettings.Type == "basic" {
		return auth.NewBasicAuth(d.Context())
	}
	return nil
}

func (d *Domain) Storage() storage.FileStorage {
	if d.entity.StorageSettings.Type == "gcs" {
		return storage.NewGcsFileStorage(d.Context(), d.entity.StorageSettings.Bucket)
	}
	return nil
}

// Context returns a namespace-wrapped context for the domain.
func (d *Domain) Context() appengine.Context {
	c, _ := appengine.Namespace(d.context, d.name)
	return c
}

func (d *Domain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := d.Context()

	s := d.Storage()
	if s == nil {
		c.Errorf("storage unconfigured for domain %v", d.name)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "not configured")
		return
	}

	a := d.Auth()
	if a != nil && !a.IsAuthorized(r) {
		// TODO(stevenle): move unauthorized handling into auth package.
		w.Header().Set("WWW-Authenticate", "Basic realm=\"Please enter a username and password\"")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "unauthorized")
		return
	}

	// TODO(stevenle): error handling.
	s.Serve(w, r)
}
