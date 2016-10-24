package main

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	capnp "zombiezen.com/go/capnproto2"
)

//Map returns a map representation of the PolicySet
func (instance PolicySet) Map() map[string]interface{} {
	returnMap := map[string]interface{}{}

	policySet := []map[string]interface{}{}

	policies, err := instance.Policies()
	CheckError(err)

	for i := 0; i < policies.Len(); i++ {
		policy := policies.At(i)
		policyMap := map[string]interface{}{}

		path, err := policy.Path()
		CheckError(err)

		policyMap["path"] = path

		verbs, err := policy.Verbs()
		verbArray := []string{}
		CheckError(err)

		for verbIndex := 0; verbIndex < verbs.Len(); verbIndex++ {
			verbValue, err := verbs.At(verbIndex)
			CheckError(err)
			verbArray = append(verbArray, verbValue)
		}

		headers, err := policy.Headers()
		CheckError(err)

		headerMap := map[string][]string{}

		for hI := 0; hI < headers.Len(); hI++ {
			key, err := headers.At(hI).Key()
			CheckError(err)
			values, err := headers.At(hI).Values()
			headerValueArray := []string{}
			for hvI := 0; hvI < values.Len(); hvI++ {
				headerValue, err := values.At(hvI)
				CheckError(err)

				headerValueArray = append(headerValueArray, headerValue)
			}
			headerMap[key] = headerValueArray
		}

		query, err := policy.Query()
		CheckError(err)

		queryMap := map[string][]string{}

		for hI := 0; hI < query.Len(); hI++ {
			key, err := query.At(hI).Key()
			CheckError(err)
			values, err := query.At(hI).Values()
			queryValueArray := []string{}
			for hvI := 0; hvI < values.Len(); hvI++ {
				queryValue, err := values.At(hvI)
				CheckError(err)

				queryValueArray = append(queryValueArray, queryValue)
			}
			queryMap[key] = queryValueArray
		}

		policySet = append(policySet, map[string]interface{}{
			"path":    path,
			"verbs":   strings.Join(verbArray, ","),
			"headers": headerMap,
			"queries": queryMap,
		})
	}

	/*

		for i := 0; i < policies.Len(); i++ {
			policy := policies.At(i)
			policyMap := map[string]interface{}{}

			path, err := policy.Path()
			CheckError(err)
			policyMap["path"] = path

			verbs, err := policy.Verbs()
			verbArray := []string{}
			CheckError(err)
			for verbIndex := 0; verbIndex < verbs.Len(); verbIndex++ {
				verbValue, err := verbs.At(verbIndex)
				CheckError(err)
				verbArray = append(verbArray, verbValue)
			}

			policyMap["verbs"] = verbArray

			returnMap["policySet"] = append(returnMap["policySet"], policyMap)
		}

	*/
	returnMap["policySet"] = policySet
	return returnMap
}

// Validate iterates through each Policy in the set.
// It returns true if any policy in its set returns true for validation.
// It returns false if every policy in its set returns false for validation.
func (instance PolicySet) Validate(request HTTPRequest) bool {
	policies, err := instance.Policies()

	if err != nil {
		panic(err)
	}

	var policyResult = policies.Len() == 0

	for i := 0; i < policies.Len(); i++ {
		policyResult = policyResult || policies.At(i).Validate(request)
		if policyResult {
			break
		}
	}

	if !policyResult {
		return false
	}

	if instance.HasDelegation() {
		log.WithFields(log.Fields{
			"hasDelegation": true,
		}).Debug("PolictSet:Validate")

		delegation, err := instance.Delegation()
		CheckError(err)
		return delegation.Validate(request)
	}

	log.WithFields(log.Fields{
		"hasDelegation": false,
	}).Debug("PolictSet:Validate")

	return true
}

// Clone creates a new PolicySet which is a clone of the instance.
// It returns the new PolicySet
func (instance PolicySet) Clone(segment *capnp.Segment) PolicySet {
	policySetBuilder := NewPolicySetBuilder()

	policies, err := instance.Policies()
	CheckError(err)

	for i := 0; i < policies.Len(); i++ {
		policyBuilder := NewPolicyBuilder()

		policy := policies.At(i)
		verbs, err := policy.Verbs()
		CheckError(err)

		for verbIndex := 0; verbIndex < verbs.Len(); verbIndex++ {
			verb, err := verbs.At(verbIndex)
			CheckError(err)
			policyBuilder = policyBuilder.WithVerb(verb)
		}

		path, err := policy.Path()
		CheckError(err)

		policyBuilder = policyBuilder.WithPath(path)

		headers, err := policy.Headers()
		CheckError(err)

		for headerIndex := 0; headerIndex < headers.Len(); i++ {
			keyValuePolicy := headers.At(headerIndex)
			key, err := keyValuePolicy.Key()
			CheckError(err)

			values, err := keyValuePolicy.Values()
			CheckError(err)
			var headerValueStrings = []string{}

			for headerValueIndex := 0; headerValueIndex < values.Len(); headerValueIndex++ {
				headerValue, err := values.At(headerValueIndex)
				CheckError(err)
				headerValueStrings = append(headerValueStrings, headerValue)
			}

			policyBuilder = policyBuilder.WithHeader(key, headerValueStrings)
		}

		queries, err := policy.Query()
		CheckError(err)

		for queryIndex := 0; queryIndex < queries.Len(); i++ {
			keyValuePolicy := queries.At(queryIndex)
			key, err := keyValuePolicy.Key()
			CheckError(err)

			values, err := keyValuePolicy.Values()
			CheckError(err)
			var queryValueStrings = []string{}

			for queryValueIndex := 0; queryValueIndex < values.Len(); queryValueIndex++ {
				queryValue, err := values.At(queryValueIndex)
				CheckError(err)
				queryValueStrings = append(queryValueStrings, queryValue)
			}

			policyBuilder = policyBuilder.WithQuery(key, queryValueStrings)
		}

		policySetBuilder = policySetBuilder.WithPolicy(policyBuilder)
	}

	return policySetBuilder.BuildPolicySet(segment)
}
