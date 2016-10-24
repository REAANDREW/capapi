package main

//PolicyJSONDto is used with the APIGatewayProxy Server in order to create delegations over HTTP.
type PolicyJSONDto struct {
	Verbs   []string            `json:"verbs"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Query   map[string][]string `json:"query"`
}