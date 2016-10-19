package tests

import (
	"testing"

	capnp "zombiezen.com/go/capnproto2"

	log "github.com/Sirupsen/logrus"
	"github.com/reaandrew/capapi/capability"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPolicySet(t *testing.T) {

	log.SetLevel(log.ErrorLevel)

	Convey("PolicySet", t, func() {

		Convey("validate returns", func() {

			Convey("true", func() {

				Convey("when there are no policies", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policySet, _ := capability.NewPolicySet(seg)

					httpRequest, _ := capability.NewHTTPRequest(seg)

					So(policySet.Validate(httpRequest), ShouldEqual, true)
				})

				Convey("when there is one policy and the policy.Validates true", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policySet, _ := capability.NewPolicySet(seg)

					policy, _ := capability.NewPolicy(seg)

					policyList, _ := capability.NewPolicy_List(seg, 1)

					policyList.Set(0, policy)

					policySet.SetPolicies(policyList)

					httpRequest, _ := capability.NewHTTPRequest(seg)
					So(policySet.Validate(httpRequest), ShouldEqual, true)
				})

				Convey("when there is two policies and only one policy.Validates true", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policySet, _ := capability.NewPolicySet(seg)

					policyList, _ := capability.NewPolicy_List(seg, 2)

					policy1, _ := capability.NewPolicy(seg)
					policy1.SetPath("/resource/A")
					policyList.Set(0, policy1)

					policy2, _ := capability.NewPolicy(seg)
					policy2.SetPath("/resource/B")
					policyList.Set(1, policy2)

					policySet.SetPolicies(policyList)

					httpRequest, _ := capability.NewHTTPRequest(seg)
					httpRequest.SetPath("/resource/B")

					So(policySet.Validate(httpRequest), ShouldEqual, true)
				})
			})

			Convey("false", func() {
				Convey("when there is one policy and the policy.Validates false", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policySet, _ := capability.NewPolicySet(seg)

					policyList, _ := capability.NewPolicy_List(seg, 1)

					policy1, _ := capability.NewPolicy(seg)
					policy1.SetPath("/resource/A")
					policyList.Set(0, policy1)

					policySet.SetPolicies(policyList)

					httpRequest, _ := capability.NewHTTPRequest(seg)
					httpRequest.SetPath("/resource/B")

					So(policySet.Validate(httpRequest), ShouldEqual, false)
				})

				Convey("when there is two policies and each policy.Validates false", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policySet, _ := capability.NewPolicySet(seg)

					policyList, _ := capability.NewPolicy_List(seg, 2)

					policy1, _ := capability.NewPolicy(seg)
					policy1.SetPath("/resource/A")
					policyList.Set(0, policy1)

					policy2, _ := capability.NewPolicy(seg)
					policy2.SetPath("/resource/B")
					policyList.Set(1, policy2)

					policySet.SetPolicies(policyList)

					httpRequest, _ := capability.NewHTTPRequest(seg)
					httpRequest.SetPath("/resource/C")

					So(policySet.Validate(httpRequest), ShouldEqual, false)
				})
			})

		})
	})

}
