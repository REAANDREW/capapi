package main

import (
	"crypto/rand"
	"encoding/base64"

	log "github.com/Sirupsen/logrus"
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

func (instance PolicySetBuilder) Build() (string, []byte) {
	msg, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	policySet, _ := NewRootPolicySet(seg)
	policyList, _ := NewPolicy_List(seg, int32(len(instance.PolicyBuilders)))

	for i := 0; i < len(instance.PolicyBuilders); i++ {
		policy := instance.PolicyBuilders[i].Build(seg)
		policyList.Set(i, policy)
	}

	policySet.SetPolicies(policyList)

	byteValue, _ := msg.Marshal()
	keyBytes := make([]byte, 64)
	_, err := rand.Read(keyBytes)
	CheckError(err)

	key := base64.StdEncoding.EncodeToString(keyBytes)
	log.WithFields(log.Fields{
		"key": key,
	}).Info("Key Generated")
	return key, byteValue
}

func NewPolicySetBuilder() PolicySetBuilder {
	return PolicySetBuilder{
		PolicyBuilders: []PolicyBuilder{},
	}
}
