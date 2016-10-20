package main

import "errors"

var (
	ErrNoAuthorizationHeader        = errors.New("No Authorization Header")
	ErrMalformedAuthorizationHeader = errors.New("Malformed Authorization Header")
	ErrNoAPIKey                     = errors.New("No API Key")
	ErrAPIKeyNotFound               = errors.New("The supplied API Key was not found")
)
