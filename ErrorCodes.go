package main

import "errors"

var (
	errNoAuthorizationHeader        = errors.New("No Authorization Header")
	errMalformedAuthorizationHeader = errors.New("Malformed Authorization Header")
	errNoAPIKey                     = errors.New("No API Key")
	errAPIKeyNotFound               = errors.New("The supplied API Key was not found")
)
