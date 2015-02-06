#!/bin/sh

set -e

# Install go deps.
go get code.google.com/p/go.crypto/bcrypt
go get code.google.com/p/goauth2/oauth
go get code.google.com/p/google-api-go-client/storage/v1
go get github.com/gorilla/rpc
