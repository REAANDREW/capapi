package main

import (
	"encoding/json"
	"io"

	log "github.com/Sirupsen/logrus"
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

//NewPolicySetFromPolicyJSONDtos create a new PolicySet from an array of PolicyJSONDto
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

		verbList, err := capnp.NewTextList(seg, int32(len(jsonPolicy.Verbs)))
		CheckError(err)

		for verbIndex, verb := range jsonPolicy.Verbs {
			verbList.Set(verbIndex, verb)
		}

		policy.SetVerbs(verbList)

		headerKeyValueList, err := NewKeyValuePolicy_List(seg, int32(len(jsonPolicy.Headers)))
		CheckError(err)

		headerCount := 0
		for headerKey, headerValues := range jsonPolicy.Headers {
			keyValuePolicy, err := NewKeyValuePolicy(seg)
			CheckError(err)

			keyValuePolicy.SetKey(headerKey)

			headerValueList, err := capnp.NewTextList(seg, int32(len(headerValues)))
			CheckError(err)

			for index, headerValue := range headerValues {
				log.WithFields(log.Fields{
					"headerKey":   headerKey,
					"headerValue": headerValue,
				}).Debug("adding headers to policy")
				headerValueList.Set(index, headerValue)
			}

			keyValuePolicy.SetValues(headerValueList)
			headerKeyValueList.Set(headerCount, keyValuePolicy)
			headerCount++
		}
		policy.SetHeaders(headerKeyValueList)

		queryKeyValueList, err := NewKeyValuePolicy_List(seg, int32(len(jsonPolicy.Query)))
		CheckError(err)

		queryCount := 0

		log.WithFields(log.Fields{
			"numberOfFields": len(jsonPolicy.Query),
		}).Debug("About to iterate over Query")

		for queryKey, queryValues := range jsonPolicy.Query {
			keyValuePolicy, err := NewKeyValuePolicy(seg)
			CheckError(err)

			keyValuePolicy.SetKey(queryKey)

			queryValueList, err := capnp.NewTextList(seg, int32(len(queryValues)))
			CheckError(err)

			log.WithFields(log.Fields{
				"queryKey": queryKey,
			}).Debug("query key")

			for index, queryValue := range queryValues {
				log.WithFields(log.Fields{
					"queryValue": queryValue,
				}).Debug("query value for key")
				queryValueList.Set(index, queryValue)
			}

			keyValuePolicy.SetValues(queryValueList)
			queryKeyValueList.Set(queryCount, keyValuePolicy)
			queryCount++
		}
		policy.SetQuery(queryKeyValueList)

		policyList.Set(index, policy)
	}

	policySet.SetPolicies(policyList)

	return policySet, nil
}
