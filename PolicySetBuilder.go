package main

import (
	capnp "zombiezen.com/go/capnproto2"
)

type PolicySetBuilder struct {
	PolicyBuilders []PolicyBuilder
}

func (instance PolicySetBuilder) WithPolicy(builder PolicyBuilder) PolicySetBuilder {
	return PolicySetBuilder{
		PolicyBuilders: append(instance.PolicyBuilders, builder),
	}
}

func (instance PolicySetBuilder) BuildPolicySet(seg *capnp.Segment) PolicySet {
	policySet, _ := NewRootPolicySet(seg)
	policyList, _ := NewPolicy_List(seg, int32(len(instance.PolicyBuilders)))

	for i := 0; i < len(instance.PolicyBuilders); i++ {
		policy := instance.PolicyBuilders[i].Build(seg)
		policyList.Set(i, policy)
	}

	policySet.SetPolicies(policyList)

	return policySet
}

func (instance PolicySetBuilder) Build() (string, []byte) {

	msg, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	instance.BuildPolicySet(seg)

	key, err := CreateKey()
	CheckError(err)

	byteValue, err := msg.Marshal()
	CheckError(err)

	return key, byteValue
}

func NewPolicySetBuilder() PolicySetBuilder {
	return PolicySetBuilder{
		PolicyBuilders: []PolicyBuilder{},
	}
}
