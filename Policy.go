package main

import (
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type validatePolicy func(policy Policy, request HTTPRequest) bool

func validateVerbs(policy Policy, request HTTPRequest) bool {
	verbs, err := policy.Verbs()

	CheckError(err)

	if verbs.Len() == 0 {
		return true
	}

	verb, err := request.Verb()

	CheckError(err)

	for i := 0; i < verbs.Len(); i++ {
		scopedVerb, _ := verbs.At(i)
		log.WithFields(log.Fields{
			"requestVerb": verb,
			"policyVerb":  scopedVerb,
		}).Debug("validating verb")
		if strings.ToUpper(verb) == strings.ToUpper(scopedVerb) {
			return true
		}
	}

	return false
}

func validateExactPath(policy Policy, request HTTPRequest) bool {
	policyPath, err := policy.Path()
	CheckError(err)

	requestPath, err := request.Path()
	CheckError(err)

	return policyPath == "" || requestPath == policyPath
}

func validateTemplatedPath(policy Policy, request HTTPRequest) bool {
	policyPath, err := policy.Path()
	CheckError(err)

	requestPath, err := request.Path()
	CheckError(err)

	r := mux.NewRouter()
	r.Path(policyPath)

	req, _ := http.NewRequest("GET", "http://localhost", nil)

	req.URL.Path = requestPath

	var routeMatch mux.RouteMatch
	return r.Match(req, &routeMatch)
}

func validateKeyValues(keyValues KeyValue_List, keyValuePolicies KeyValuePolicy_List) bool {
	log.WithFields(log.Fields{
		"keyValues:Length":        keyValues.Len(),
		"keyValuePolicies:Length": keyValuePolicies.Len(),
	}).Debug("validateQueryValues:enter")

	for i := 0; i < keyValues.Len(); i++ {
		valid := false
		log.WithFields(log.Fields{
			"numberOfPolicies": keyValuePolicies.Len(),
		}).Debug("Looping query policies")
		for j := 0; j < keyValuePolicies.Len(); j++ {
			req := keyValues.At(i)

			reqKey, err := req.Key()
			CheckError(err)

			reqKeyValue, err := req.Value()
			CheckError(err)

			policy := keyValuePolicies.At(i)

			policyKey, err := policy.Key()
			CheckError(err)

			policyValues, err := policy.Values()
			CheckError(err)

			if reqKey == policyKey {
				if policyValues.Len() == 0 {
					valid = true
				} else {
					for k := 0; k < policyValues.Len(); k++ {
						policyKeyValue, err := policyValues.At(k)
						CheckError(err)

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
		log.WithFields(log.Fields{
			"valid": valid,
		}).Debug("Exiting querystring validation")
		if !valid {
			return false
		}
	}
	return true
}

func validateHeaderValues(keyValues KeyValue_List, keyValuePolicies KeyValuePolicy_List) bool {

	log.WithFields(log.Fields{
		"keyValues:Length":        keyValues.Len(),
		"keyValuePolicies:Length": keyValuePolicies.Len(),
	}).Debug("validateHeaderValues:enter")

	for i := 0; i < keyValues.Len(); i++ {
		valid := false
		for j := 0; j < keyValuePolicies.Len(); j++ {
			req := keyValues.At(i)

			reqKey, err := req.Key()
			CheckError(err)

			reqKeyValue, err := req.Value()
			CheckError(err)

			policy := keyValuePolicies.At(j)

			policyKey, err := policy.Key()
			CheckError(err)

			policyValues, err := policy.Values()
			CheckError(err)

			log.WithFields(log.Fields{
				"policyKey": policyKey,
				"reqKey":    reqKey,
			}).Debug("validateHeaderValues:checking")

			if reqKey == policyKey {
				if policyValues.Len() == 0 {
					valid = true
				} else {
					for k := 0; k < policyValues.Len(); k++ {
						policyKeyValue, err := policyValues.At(k)
						CheckError(err)

						log.WithFields(log.Fields{
							"policyKeyValue": policyKeyValue,
							"reqKeyValue":    reqKeyValue,
						}).Debug("validateHeaderValues:checking")

						if reqKeyValue == policyKeyValue {
							valid = true
							break
						}
					}
				}
				log.WithFields(log.Fields{
					"valid": valid,
				}).Debug("validateHeaderValues")
				if !valid {
					return false
				}
			}
		}
	}
	return true
}
func validateHeaders(policy Policy, request HTTPRequest) bool {
	reqHeaders, err := request.Headers()
	CheckError(err)

	policyHeaders, err := policy.Headers()
	CheckError(err)

	return validateHeaderValues(reqHeaders, policyHeaders)
}

func validateQuery(policy Policy, request HTTPRequest) bool {
	reqQuery, err := request.Query()
	CheckError(err)

	policyQuery, err := policy.Query()
	CheckError(err)

	return validateKeyValues(reqQuery, policyQuery)
}

func validatePath(policy Policy, request HTTPRequest) bool {
	return validateExactPath(policy, request) ||
		validateTemplatedPath(policy, request)
}

//Validate validates request against its the Policy state and returns the result.
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
	}).Debug("validate")

	return verbResult && pathResult && headersResult && queryResult
}
