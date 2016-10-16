package main

import (
	"net/http"
	"strings"
)

func parseAuthorization(request *http.Request) (string, error) {

	header := request.Header.Get("Authorization")
	if header == "" {
		return "", errNoAuthorizationHeader
	}

	if !strings.Contains(header, "Bearer") {
		return "", errMalformedAuthorizationHeader
	}

	split := strings.Split(header, " ")

	if len(split) != 2 {
		return "", errNoAPIKey
	}

	return split[1], nil
}
