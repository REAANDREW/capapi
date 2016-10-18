package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

/*
import (
	"github.com/gorilla/mux"
)
*/

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

func validatePath(policy Policy, request HTTPRequest) bool {
	return validateExactPath(policy, request) ||
		validateTemplatedPath(policy, request)
}

func (instance Policy) validate(request HTTPRequest) bool {
	return validateVerbs(instance, request) &&
		validatePath(instance, request)
}
