package main

import (
	"strings"

	capnp "zombiezen.com/go/capnproto2"
)

//TextListToArray returns a simple string array of the TextList
func TextListToArray(instance capnp.TextList) []string {
	returnArray := []string{}
	for i := 0; i < instance.Len(); i++ {
		value, err := instance.At(i)
		CheckError(err)
		returnArray = append(returnArray, value)
	}
	return returnArray
}

//Map returns a map representation of the KeyValuePolicy
func (instance KeyValuePolicy_List) Map() map[string][]string {

	returnMap := map[string][]string{}

	for i := 0; i < instance.Len(); i++ {
		key, err := instance.At(i).Key()
		CheckError(err)
		values, err := instance.At(i).Values()
		valueArray := []string{}
		for hvI := 0; hvI < values.Len(); hvI++ {
			value, err := values.At(hvI)
			CheckError(err)

			valueArray = append(valueArray, value)
		}
		returnMap[key] = valueArray
	}

	return returnMap

}

//Map returns a map representation of the PolicySet
func (instance PolicySet) Map() map[string]interface{} {
	returnMap := map[string]interface{}{}

	policySet := []map[string]interface{}{}

	policies, err := instance.Policies()
	CheckError(err)

	for i := 0; i < policies.Len(); i++ {
		policy := policies.At(i)

		path, err := policy.Path()
		CheckError(err)

		verbs, err := policy.Verbs()
		CheckError(err)

		headers, err := policy.Headers()
		CheckError(err)

		query, err := policy.Query()
		CheckError(err)

		policySet = append(policySet, map[string]interface{}{
			"path":    path,
			"verbs":   strings.Join(TextListToArray(verbs), ","),
			"headers": headers.Map(),
			"queries": query.Map(),
		})
	}

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
		delegation, err := instance.Delegation()
		CheckError(err)
		return delegation.Validate(request)
	}

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
		policyBuilder = policyBuilder.WithVerbs(TextListToArray(verbs))

		path, err := policy.Path()
		CheckError(err)
		policyBuilder = policyBuilder.WithPath(path)

		headers, err := policy.Headers()
		CheckError(err)
		policyBuilder = policyBuilder.WithHeaders(headers.Map())

		queries, err := policy.Query()
		CheckError(err)
		policyBuilder = policyBuilder.WithQueries(queries.Map())

		policySetBuilder = policySetBuilder.WithPolicy(policyBuilder)
	}

	return policySetBuilder.BuildPolicySet(segment)
}
