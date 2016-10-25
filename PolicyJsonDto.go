package main

import (
	"encoding/json"
	"io"

	capnp "zombiezen.com/go/capnproto2"
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

//TextListFromArray creates a new TextList from the given input string array
//TODO: Instead of checking the errors let them propagate up the chain
func TextListFromArray(input []string) (capnp.TextList, error) {
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	list, err := capnp.NewTextList(seg, int32(len(input)))
	CheckError(err)

	for index, value := range input {
		list.Set(index, value)
	}

	return list, nil
}

//NewPolicySetFromPolicyJSONDtos create a new PolicySet from an array of PolicyJSONDto
//TODO: Instead of checking the errors let them propagate up the chain
func NewPolicySetFromPolicyJSONDtos(policies []PolicyJSONDto) (PolicySet, error) {

	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		panic(err)
	}
	policySet, err := NewPolicySet(seg)
	CheckError(err)

	policyList, err := NewPolicy_List(seg, int32(len(policies)))
	CheckError(err)

	for index, jsonPolicy := range policies {
		policy, err := NewPolicy(seg)
		CheckError(err)

		policy.SetPath(jsonPolicy.Path)

		verbList, err := TextListFromArray(jsonPolicy.Verbs)
		CheckError(err)
		policy.SetVerbs(verbList)

		headerKeyValueList, err := KeyValuePolicyListFromMap(jsonPolicy.Headers)
		CheckError(err)
		policy.SetHeaders(headerKeyValueList)

		queryKeyValueList, err := KeyValuePolicyListFromMap(jsonPolicy.Query)
		CheckError(err)
		policy.SetQuery(queryKeyValueList)

		policyList.Set(index, policy)
	}

	policySet.SetPolicies(policyList)

	return policySet, nil
}
