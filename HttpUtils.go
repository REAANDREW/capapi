package main

import (
	"net/http"
	"strings"
)

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
