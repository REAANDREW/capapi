package capability

import (
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/reaandrew/capapi/core"
)

type validatePolicy func(policy Policy, request HTTPRequest) bool

func validateVerbs(policy Policy, request HTTPRequest) bool {
	verbs, err := policy.Verbs()

	core.CheckError(err)

	if verbs.Len() == 0 {
		return true
	}

	verb, err := request.Verb()

	core.CheckError(err)

	for i := 0; i < verbs.Len(); i++ {
		scopedVerb, _ := verbs.At(i)
		log.WithFields(log.Fields{
			"requestVerb": verb,
			"policyVerb":  scopedVerb,
		}).Debug("validating verb")
		if verb == scopedVerb {
			return true
		}
	}

	return false
}

func validateExactPath(policy Policy, request HTTPRequest) bool {
	policyPath, err := policy.Path()
	core.CheckError(err)

	requestPath, err := request.Path()
	core.CheckError(err)

	return requestPath == policyPath
}

func validateTemplatedPath(policy Policy, request HTTPRequest) bool {
	policyPath, err := policy.Path()
	core.CheckError(err)

	requestPath, err := request.Path()
	core.CheckError(err)

	r := mux.NewRouter()
	r.Path(policyPath)

	req, _ := http.NewRequest("GET", "http://localhost", nil)

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
			core.CheckError(err)

			reqKeyValue, err := req.Value()
			core.CheckError(err)

			policy := keyValuePolicies.At(i)

			policyKey, err := policy.Key()
			core.CheckError(err)

			policyValues, err := policy.Values()
			core.CheckError(err)

			if reqKey == policyKey {
				if policyValues.Len() == 0 {
					valid = true
				} else {
					for k := 0; k < policyValues.Len(); k++ {
						policyKeyValue, err := policyValues.At(k)
						core.CheckError(err)

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
	core.CheckError(err)

	policyHeaders, err := policy.Headers()
	core.CheckError(err)

	return validateKeyValues(reqHeaders, policyHeaders)
}

func validateQuery(policy Policy, request HTTPRequest) bool {
	reqQuery, err := request.Query()
	core.CheckError(err)

	policyQuery, err := policy.Query()
	core.CheckError(err)

	return validateKeyValues(reqQuery, policyQuery)
}

func validatePath(policy Policy, request HTTPRequest) bool {
	return validateExactPath(policy, request) ||
		validateTemplatedPath(policy, request)
}

func (instance Policy) Validate(request HTTPRequest) bool {
	verbResult := validateVerbs(instance, request)
	pathResult := validatePath(instance, request)
	headersResult := validateHeaders(instance, request)
	queryResult := validateQuery(instance, request)

	log.WithFields(log.Fields{
		"verbResult":   strconv.FormatBool(verbResult),
		"pathResult":   strconv.FormatBool(pathResult),
		"headerResult": strconv.FormatBool(headersResult),
		"queryResult":  strconv.FormatBool(queryResult),
	}).Info("validate")

	return verbResult && pathResult && headersResult && queryResult
}
