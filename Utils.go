package main

import (
	"net/http"
	"strings"
)

func checkError(err error) {

	//Stay verbose for the time being
	if err != nil {
		panic(err)
	}
}

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
