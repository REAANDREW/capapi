package main

//PolicyJSONDto is used with the APIGatewayProxy Server in order to create delegations over HTTP.
type PolicyJSONDto struct {
	Verbs []string `json:"verbs"`
}
