package main

import (
	capnp "zombiezen.com/go/capnproto2"
)

//PolicySetBuilder is used to build PolicySets with a fluent interface.
type PolicySetBuilder struct {
	PolicyBuilders []PolicyBuilder
}

//WithPolicy takes a PolicyBuilder and adds it to the collection of PolicyBuilders.
func (instance PolicySetBuilder) WithPolicy(builder PolicyBuilder) PolicySetBuilder {
	return PolicySetBuilder{
		PolicyBuilders: append(instance.PolicyBuilders, builder),
	}
}

//BuildPolicySet takes a message segment, iterates over the PolicyBuilders.
func (instance PolicySetBuilder) BuildPolicySet() PolicySet {
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	policySet, _ := NewRootPolicySet(seg)
	policyList, _ := NewPolicy_List(seg, int32(len(instance.PolicyBuilders)))

	for i := 0; i < len(instance.PolicyBuilders); i++ {
		policy := instance.PolicyBuilders[i].Build(seg)
		policyList.Set(i, policy)
	}

	policySet.SetPolicies(policyList)

	return policySet
}

//Build returns a string key and also the byte representation of a built PolicySet.
func (instance PolicySetBuilder) Build() (string, []byte) {

	msg, _, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	policySet := instance.BuildPolicySet()

	key, err := CreateKey()
	CheckError(err)

	msg.SetRootPtr(policySet.ToPtr())

	byteValue, err := msg.Marshal()
	CheckError(err)

	return key, byteValue
}

//NewPolicySetBuilder creates, initializes and returns a new PolicySetBuilder.
func NewPolicySetBuilder() PolicySetBuilder {
	return PolicySetBuilder{
		PolicyBuilders: []PolicyBuilder{},
	}
}
