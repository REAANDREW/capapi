package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type validatePolicy func(policy Policy, request HTTPRequest) bool

func validateVerbs(policy Policy, request HTTPRequest) bool {
	verbs, err := policy.Verbs()

	checkError(err)

	if verbs.Len() == 0 {
		return true
	}

	verb, err := request.Verb()

	checkError(err)

	for i := 0; i < verbs.Len(); i++ {
		scopedVerb, _ := verbs.At(i)
		if verb == scopedVerb {
			return true
		}
	}

	return false
}

func validateExactPath(policy Policy, request HTTPRequest) bool {
	policyPath, err := policy.Path()
	checkError(err)

	requestPath, err := request.Path()
	checkError(err)

	return requestPath == policyPath
}

func validateTemplatedPath(policy Policy, request HTTPRequest) bool {
	policyPath, err := policy.Path()
	checkError(err)

	requestPath, err := request.Path()
	checkError(err)

	r := mux.NewRouter()
	r.Path(policyPath)

	req, _ := http.NewRequest("GET", "http://localhost:34567", nil)

	req.URL.Path = requestPath

	var routeMatch mux.RouteMatch
	return r.Match(req, &routeMatch)
}

func validateKeyValues(keyValues KeyValue_List, keyValuePolicies KeyValuePolicy_List) bool {
	for i := 0; i < keyValues.Len(); i++ {
		valid := false
		for j := 0; j < keyValuePolicies.Len(); j++ {
			req := keyValues.At(i)

			reqKey, err := req.Key()
			checkError(err)

			reqKeyValue, err := req.Value()
			checkError(err)

			policy := keyValuePolicies.At(i)

			policyKey, err := policy.Key()
			checkError(err)

			policyValues, err := policy.Values()
			checkError(err)

			if reqKey == policyKey {
				if policyValues.Len() == 0 {
					valid = true
				} else {
					for k := 0; k < policyValues.Len(); k++ {
						policyKeyValue, err := policyValues.At(k)
						checkError(err)

						if reqKeyValue == policyKeyValue {
							valid = true
						}
					}
				}
				if valid {
					break
				}
			}
		}
		if !valid {
			return false
		}
	}
	return true
}

func validateHeaders(policy Policy, request HTTPRequest) bool {
	reqHeaders, err := request.Headers()
	checkError(err)

	policyHeaders, err := policy.Headers()
	checkError(err)

	return validateKeyValues(reqHeaders, policyHeaders)
}

func validateQuery(policy Policy, request HTTPRequest) bool {
	reqQuery, err := request.Query()
	checkError(err)

	policyQuery, err := policy.Query()
	checkError(err)

	return validateKeyValues(reqQuery, policyQuery)
}

func validatePath(policy Policy, request HTTPRequest) bool {
	return validateExactPath(policy, request) ||
		validateTemplatedPath(policy, request)
}

func (instance Policy) validate(request HTTPRequest) bool {
	return validateVerbs(instance, request) &&
		validatePath(instance, request) &&
		validateHeaders(instance, request) &&
		validateQuery(instance, request)
}
