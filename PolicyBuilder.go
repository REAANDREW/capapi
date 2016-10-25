package main

import capnp "zombiezen.com/go/capnproto2"

//PolicyBuilder is a builder to create a new Policy
type PolicyBuilder struct {
	Path    string
	Verbs   []string
	Headers map[string][]string
	Query   map[string][]string
}

//NewPolicyBuilder creates, initializes and returns a PolicyBuilder.
func NewPolicyBuilder() PolicyBuilder {
	return PolicyBuilder{
		Verbs:   []string{},
		Headers: map[string][]string{},
		Query:   map[string][]string{},
	}
}

//WithPath sets the path and returns a PolicyBuilder.
func (instance PolicyBuilder) WithPath(path string) PolicyBuilder {
	return PolicyBuilder{
		Path:    path,
		Verbs:   instance.Verbs,
		Headers: instance.Headers,
		Query:   instance.Query,
	}
}

//WithVerb adds a verb to the collection of the policy and returns a PolicyBuilder.
func (instance PolicyBuilder) WithVerb(verb string) PolicyBuilder {
	return PolicyBuilder{
		Path:    instance.Path,
		Verbs:   append(instance.Verbs, verb),
		Headers: instance.Headers,
		Query:   instance.Query,
	}
}

//WithVerbs adds each of the verbs to the collection of the policy and returns a PolicyBuilder.
func (instance PolicyBuilder) WithVerbs(verbs []string) PolicyBuilder {
	var returnBuilder = instance

	for _, verb := range verbs {
		returnBuilder = returnBuilder.WithVerb(verb)
	}

	return returnBuilder
}

//WithHeader adds a Header and permitted values to the policy and returns a PolicyBuilder.
func (instance PolicyBuilder) WithHeader(key string, values []string) PolicyBuilder {
	headers := instance.Headers
	headers[key] = values
	return PolicyBuilder{
		Path:    instance.Path,
		Verbs:   instance.Verbs,
		Headers: headers,
		Query:   instance.Query,
	}
}

//WithHeaders adds the headers to the policy and returns a PolicyBuilder
func (instance PolicyBuilder) WithHeaders(headers map[string][]string) PolicyBuilder {
	var returnBuilder = instance

	for key, values := range headers {
		returnBuilder = returnBuilder.WithHeader(key, values)
	}

	return returnBuilder
}

//WithQuery adds a QueryString key and permitted values to the policy and returns a PolicyBuilder.
func (instance PolicyBuilder) WithQuery(key string, values []string) PolicyBuilder {
	query := instance.Query
	query[key] = values
	return PolicyBuilder{
		Path:    instance.Path,
		Verbs:   instance.Verbs,
		Headers: instance.Headers,
		Query:   query,
	}
}

//WithQueries adds the queries to the policy and returns a PolicyBuilder
func (instance PolicyBuilder) WithQueries(queries map[string][]string) PolicyBuilder {
	var returnBuilder = instance

	for key, values := range queries {
		returnBuilder = returnBuilder.WithQuery(key, values)
	}

	return returnBuilder
}

//Build takes a message segment, builds and returns a Policy.
func (instance PolicyBuilder) Build(seg *capnp.Segment) Policy {
	policy, _ := NewPolicy(seg)

	verbList, _ := capnp.NewTextList(seg, int32(len(instance.Verbs)))
	for i := 0; i < len(instance.Verbs); i++ {
		verbList.Set(i, instance.Verbs[i])
	}
	policy.SetVerbs(verbList)

	headerList, _ := KeyValuePolicyListFromMap(instance.Headers)
	policy.SetHeaders(headerList)

	queryList, _ := KeyValuePolicyListFromMap(instance.Query)
	policy.SetQuery(queryList)

	policy.SetPath(instance.Path)

	return policy
}
