package main

import capnp "zombiezen.com/go/capnproto2"

type PolicyBuilder struct {
	Path    string
	Verbs   []string
	Headers map[string][]string
	Query   map[string][]string
}

func NewPolicyBuilder() PolicyBuilder {
	return PolicyBuilder{
		Verbs:   []string{},
		Headers: map[string][]string{},
		Query:   map[string][]string{},
	}
}

func (instance PolicyBuilder) WithPath(path string) PolicyBuilder {
	return PolicyBuilder{
		Path:    path,
		Verbs:   instance.Verbs,
		Headers: instance.Headers,
		Query:   instance.Query,
	}
}

func (instance PolicyBuilder) WithVerb(verb string) PolicyBuilder {
	return PolicyBuilder{
		Path:    instance.Path,
		Verbs:   append(instance.Verbs, verb),
		Headers: instance.Headers,
		Query:   instance.Query,
	}
}

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

func (instance PolicyBuilder) Build(seg *capnp.Segment) Policy {
	policy, _ := NewPolicy(seg)

	verbList, _ := capnp.NewTextList(seg, int32(len(instance.Verbs)))
	for i := 0; i < len(instance.Verbs); i++ {
		verbList.Set(i, instance.Verbs[i])
	}
	policy.SetVerbs(verbList)

	headerList, _ := NewKeyValuePolicy_List(seg, int32(len(instance.Headers)))
	count := 0
	for key, value := range instance.Headers {
		headerPolicy, _ := NewKeyValuePolicy(seg)
		headerPolicy.SetKey(key)

		headerValueList, _ := capnp.NewTextList(seg, int32(len(value)))
		for i := 0; i < len(value); i++ {
			headerValueList.Set(i, value[i])
		}
		headerPolicy.SetValues(headerValueList)

		headerList.Set(count, headerPolicy)
		count++
	}
	policy.SetHeaders(headerList)

	queryList, _ := NewKeyValuePolicy_List(seg, int32(len(instance.Query)))
	count = 0
	for key, value := range instance.Query {
		queryPolicy, _ := NewKeyValuePolicy(seg)
		queryPolicy.SetKey(key)

		queryValueList, _ := capnp.NewTextList(seg, int32(len(value)))
		for i := 0; i < len(value); i++ {
			queryValueList.Set(i, value[i])
		}
		queryPolicy.SetValues(queryValueList)

		queryList.Set(count, queryPolicy)
		count++
	}
	policy.SetQuery(queryList)

	policy.SetPath(instance.Path)

	return policy
}
