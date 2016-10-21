package main

import (
	"net/http"
	"strings"
)

//ParseAuthorization returns the Bearer Authorization Token value
//Returns the authorization token value if present
//Returns an error if it is not present
func ParseAuthorization(request *http.Request) (string, error) {

	header := request.Header.Get("Authorization")
	if header == "" {
		return "", ErrNoAuthorizationHeader
	}

	if !strings.Contains(header, "Bearer") {
		return "", ErrMalformedAuthorizationHeader
	}

	split := strings.Split(header, " ")

	if len(split) != 2 {
		return "", ErrNoAPIKey
	}

	return split[1], nil
}
