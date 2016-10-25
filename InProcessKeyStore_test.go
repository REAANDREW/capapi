package main

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInProcessKeyStore(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	keyStore := CreateInProcKeyStore()

	Convey("Returns", t, func() {

		Convey("err when capability key is not found", func() {

			_, err := keyStore.Get("talula")

			So(err, ShouldEqual, ErrAPIKeyNotFound)
		})

		Convey("value when key is found", func() {

			var expectedKey = "fubar"
			var expectedValue = []byte{1, 2, 3}

			keyStore.Set(expectedKey, expectedValue)
			value, _ := keyStore.Get(expectedKey)

			So(value, ShouldResemble, expectedValue)
		})
	})

	Convey("Delegate", t, func() {

		key, _ := CreateKey()
		policySet := NewPolicySetBuilder().
			WithPolicy(NewPolicyBuilder().WithVerbs([]string{"GET", "POST", "PUT"})).
			BuildPolicySet()
		keyStore.Set(key, policySet.Bytes())

		delegatedKey, _ := CreateKey()
		delegatedPolicySet := NewPolicySetBuilder().
			WithPolicy(NewPolicyBuilder().WithVerbs([]string{"GET", "POST"})).
			BuildPolicySet()

		Convey("Delegate should succeed", func() {
			err := keyStore.Delegate(key, delegatedKey, delegatedPolicySet)
			So(err, ShouldBeNil)

			Convey("The delegation will be stored against the supplied key", func() {
				result, _ := keyStore.Get(delegatedKey)
				So(result, ShouldNotBeNil)

				Convey("The delegation is wrapped with the parent delegation", func() {
					resultPolicySet := PolicySetFromBytes(result)
					So(resultPolicySet.NumberOfPoliciesEquals(1), ShouldBeTrue)

					So(resultPolicySet.Policy(0).HasVerb("GET"), ShouldBeTrue)
					So(resultPolicySet.Policy(0).HasVerb("POST"), ShouldBeTrue)
					So(resultPolicySet.Policy(0).HasVerb("PUT"), ShouldBeTrue)

					Convey("The delegation is attached to the parent delegation", func() {
						delegated, _ := resultPolicySet.Delegation()
						So(delegated.Policy(0).HasVerb("GET"), ShouldBeTrue)
						So(delegated.Policy(0).HasVerb("POST"), ShouldBeTrue)
						So(delegated.Policy(0).HasVerb("PUT"), ShouldBeFalse)
					})
				})
			})

			Convey("Revoke should succeed", func() {
				err := keyStore.Revoke(delegatedKey)
				So(err, ShouldBeNil)
				_, err = keyStore.Get(delegatedKey)
				So(err, ShouldEqual, ErrAPIKeyNotFound)
			})

			Convey("Revoking a delegation also revokes the child delegation", func() {
				err := keyStore.Revoke(key)
				So(err, ShouldBeNil)
				_, err = keyStore.Get(delegatedKey)
				So(err, ShouldEqual, ErrAPIKeyNotFound)
				_, err = keyStore.Get(key)
				So(err, ShouldEqual, ErrAPIKeyNotFound)
			})

			Convey("Revoking a delagtion with multiple levels of delegation", func() {
				furtherDelegatedKey, _ := CreateKey()
				furtherDelegatedPolicySet := NewPolicySetBuilder().
					WithPolicy(NewPolicyBuilder().WithVerb("GET")).
					BuildPolicySet()
				keyStore.Delegate(delegatedKey, furtherDelegatedKey, furtherDelegatedPolicySet)

				err := keyStore.Revoke(key)
				So(err, ShouldBeNil)
				_, err = keyStore.Get(furtherDelegatedKey)
				So(err, ShouldEqual, ErrAPIKeyNotFound)
				_, err = keyStore.Get(delegatedKey)
				So(err, ShouldEqual, ErrAPIKeyNotFound)
				_, err = keyStore.Get(key)
				So(err, ShouldEqual, ErrAPIKeyNotFound)

			})
		})
	})
}
