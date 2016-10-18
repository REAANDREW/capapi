package main

import (
	"fmt"
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

func validateHeaders(policy Policy, request HTTPRequest) bool {
	reqHeaders, err := request.Headers()
	checkError(err)

	policyHeaders, err := policy.Headers()
	checkError(err)

	for i := 0; i < reqHeaders.Len(); i++ {
		valid := false
		for j := 0; j < policyHeaders.Len(); j++ {
			req := reqHeaders.At(i)

			reqKey, err := req.Key()
			checkError(err)

			reqKeyValue, err := req.Value()
			checkError(err)

			policy := policyHeaders.At(i)

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

	fmt.Println(fmt.Sprintf("Returning true"))
	return true
}

func validatePath(policy Policy, request HTTPRequest) bool {
	return validateExactPath(policy, request) ||
		validateTemplatedPath(policy, request)
}

func (instance Policy) validate(request HTTPRequest) bool {
	return validateVerbs(instance, request) &&
		validatePath(instance, request) &&
		validateHeaders(instance, request)
}
