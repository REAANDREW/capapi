package main

import (
	"testing"

	capnp "zombiezen.com/go/capnproto2"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPolicySet(t *testing.T) {

	log.SetLevel(log.ErrorLevel)

	Convey("PolicySet", t, func() {

		Convey("validates", func() {

			Convey("true", func() {

				Convey("when there are no policies", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
					policySet, _ := NewPolicySet(seg)

					httpRequest, _ := NewHTTPRequest(seg)

					So(policySet.validate(httpRequest), ShouldEqual, true)
				})

				Convey("when there is one policy and the policy validates true", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policySet, _ := NewPolicySet(seg)

					policy, _ := NewPolicy(seg)

					policyList, _ := NewPolicy_List(seg, 1)

					policyList.Set(0, policy)

					policySet.SetPolicies(policyList)

					httpRequest, _ := NewHTTPRequest(seg)
					So(policySet.validate(httpRequest), ShouldEqual, true)
				})

				Convey("when there is two policies and only one policy validates true", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policySet, _ := NewPolicySet(seg)

					policyList, _ := NewPolicy_List(seg, 2)

					policy1, _ := NewPolicy(seg)
					policy1.SetPath("/resource/A")
					policyList.Set(0, policy1)

					policy2, _ := NewPolicy(seg)
					policy2.SetPath("/resource/B")
					policyList.Set(1, policy2)

					policySet.SetPolicies(policyList)

					httpRequest, _ := NewHTTPRequest(seg)
					httpRequest.SetPath("/resource/B")

					So(policySet.validate(httpRequest), ShouldEqual, true)
				})
			})

			Convey("false", func() {
				Convey("when there is one policy and the policy validates false", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policySet, _ := NewPolicySet(seg)

					policyList, _ := NewPolicy_List(seg, 1)

					policy1, _ := NewPolicy(seg)
					policy1.SetPath("/resource/A")
					policyList.Set(0, policy1)

					policySet.SetPolicies(policyList)

					httpRequest, _ := NewHTTPRequest(seg)
					httpRequest.SetPath("/resource/B")

					So(policySet.validate(httpRequest), ShouldEqual, false)
				})

				Convey("when there is two policies and each policy validates false", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policySet, _ := NewPolicySet(seg)

					policyList, _ := NewPolicy_List(seg, 2)

					policy1, _ := NewPolicy(seg)
					policy1.SetPath("/resource/A")
					policyList.Set(0, policy1)

					policy2, _ := NewPolicy(seg)
					policy2.SetPath("/resource/B")
					policyList.Set(1, policy2)

					policySet.SetPolicies(policyList)

					httpRequest, _ := NewHTTPRequest(seg)
					httpRequest.SetPath("/resource/C")

					So(policySet.validate(httpRequest), ShouldEqual, false)
				})
			})

		})
	})

}
