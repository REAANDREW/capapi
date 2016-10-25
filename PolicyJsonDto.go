package main

import (
	"encoding/json"
	"io"
)

//PolicyJSONDto is used with the APIGatewayProxy Server in order to create delegations over HTTP.
type PolicyJSONDto struct {
	Verbs   []string            `json:"verbs"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Query   map[string][]string `json:"query"`
}

//DecodePolicyJSONDtos decodes the DTOs from a byte stream
func DecodePolicyJSONDtos(body io.ReadCloser) []PolicyJSONDto {
	var policies []PolicyJSONDto

	decoder := json.NewDecoder(body)
	err := decoder.Decode(&policies)

	CheckError(err)

	return policies
}
