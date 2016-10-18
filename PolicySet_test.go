package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPolicySet(t *testing.T) {

	Convey("PolicySet", t, func() {

		Convey("validates", func() {

			Convey("true", func() {

				Convey("when there are no policies", func() {

				})

				Convey("when there is one policy and the policy validates true", func() {

				})

				Convey("when there is two policies and only one policy validates true", func() {

				})
			})

			Convey("false", func() {
				Convey("when there is one policy and the policy validates false", func() {

				})

				Convey("when there is two policies and each policy validates false", func() {

				})
			})

		})
	})

}
