package http

import (
	"net/http"
	"strings"

	"github.com/reaandrew/capapi/core"
)

func ParseAuthorization(request *http.Request) (string, error) {

	header := request.Header.Get("Authorization")
	if header == "" {
		return "", core.ErrNoAuthorizationHeader
	}

	if !strings.Contains(header, "Bearer") {
		return "", core.ErrMalformedAuthorizationHeader
	}

	split := strings.Split(header, " ")

	if len(split) != 2 {
		return "", core.ErrNoAPIKey
	}

	return split[1], nil
}
