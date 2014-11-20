package auth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"appengine"
	"appengine/datastore"
	"code.google.com/p/go.crypto/bcrypt"
)

var (
	ErrInvalidUsername = errors.New("auth: invalid username")
)

type Authenticator interface {
	IsAuthorized(r *http.Request) bool
}

// BasicAuth uses HTTP basic auth to authenticate and authorize a user.
type BasicAuth struct {
}

type BasicAuthUser struct {
	Username     string
	PasswordHash []byte
}

const EntityBasicAuthUser = "BasicAuthUser"

func NewBasicAuth() *BasicAuth {
	return &BasicAuth{}
}

func (b *BasicAuth) IsAuthorized(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Basic ") {
		return false
	}

	b64Value := authHeader[6:]
	decodedValue, err := base64.StdEncoding.DecodeString(b64Value)
	if err != nil {
		return false
	}

	userPass := string(decodedValue)
	i := strings.IndexRune(userPass, ':')
	if i < 0 {
		return false
	}

	username := userPass[:i]
	password := userPass[i+1:]
	c := appengine.NewContext(r)
	return b.Authenticate(c, username, password)
}

func (b *BasicAuth) AddUser(c appengine.Context, username string, password string) error {
	// Only allow alphanumeric chars in username.
	matched, _ := regexp.MatchString("[A-Za-z0-9]+", username)
	if !matched {
		return ErrInvalidUsername
	}

	// Use bcrypt for hashing passwords.
	// http://stackoverflow.com/questions/18545676
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &BasicAuthUser{
		Username:     username,
		PasswordHash: hash,
	}
	key := b.keyForUser(c, username)
	_, err = datastore.Put(c, key, user)
	return err
}

func (b *BasicAuth) Authenticate(c appengine.Context, username string, password string) bool {
	key := b.keyForUser(c, username)
	user := new(BasicAuthUser)
	err := datastore.Get(c, key, user)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	return err == nil
}

func (b *BasicAuth) keyForUser(c appengine.Context, username string) *datastore.Key {
	return datastore.NewKey(c, EntityBasicAuthUser, username, 0 /* intID */, nil /* parent */)
}
