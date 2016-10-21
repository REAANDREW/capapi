package main

import "errors"

var (
	//ErrNoAuthorizationHeader when no authorization header is present
	ErrNoAuthorizationHeader = errors.New("No Authorization Header")
	//ErrMalformedAuthorizationHeader when the authorization header is malformed.  It needs to contain Bearer Authorization
	ErrMalformedAuthorizationHeader = errors.New("Malformed Authorization Header")
	//ErrNoAPIKey when no api key has been supplied
	ErrNoAPIKey = errors.New("No API Key")
	//ErrAPIKeyNotFound when the supplied API Key is not present in the key store
	ErrAPIKeyNotFound = errors.New("The supplied API Key was not found")
)
