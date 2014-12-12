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
	context appengine.Context
}

type BasicAuthUser struct {
	Username     string
	PasswordHash []byte
}

const EntityBasicAuthUser = "BasicAuthUser"

func NewBasicAuth(c appengine.Context) *BasicAuth {
	return &BasicAuth{c}
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
	return b.Authenticate(username, password)
}

func (b *BasicAuth) AddUser(username string, password string) error {
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
	key := b.keyForUser(username)
	_, err = datastore.Put(b.context, key, user)
	return err
}

func (b *BasicAuth) Authenticate(username string, password string) bool {
	key := b.keyForUser(username)
	user := new(BasicAuthUser)
	err := datastore.Get(b.context, key, user)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	return err == nil
}

func (b *BasicAuth) keyForUser(username string) *datastore.Key {
	return datastore.NewKey(b.context, EntityBasicAuthUser, username, 0 /* intID */, nil /* parent */)
}
