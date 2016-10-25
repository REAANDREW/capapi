package main

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInProcessKeyStore(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	Convey("Returns", t, func() {
		keyStore := InProcessKeyStore{
			Keys: map[string][]byte{},
		}

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
			WithPolicy(NewPolicyBuilder().WithVerbs([]string{"GET", "POST"})).
			BuildPolicySet()

		store := InProcessKeyStore{
			Keys: map[string][]byte{
				key: policySet.Bytes(),
			},
		}

		delegatedKey, _ := CreateKey()
		delegatedPolicySet := NewPolicySetBuilder().
			WithPolicy(NewPolicyBuilder().WithVerb("GET")).
			BuildPolicySet()

		Convey("Delegate should succeed", func() {
			err := store.Delegate(key, delegatedKey, delegatedPolicySet)
			So(err, ShouldBeNil)

			Convey("The delegation will be stored against the supplied key", func() {
				result, _ := store.Get(delegatedKey)
				So(result, ShouldNotBeNil)

				Convey("The delegation is wrapped with the parent delegation", func() {
					resultPolicySet := PolicySetFromBytes(result)
					So(resultPolicySet.NumberOfPoliciesEquals(1), ShouldBeTrue)

					So(resultPolicySet.Policy(0).HasVerb("GET"), ShouldBeTrue)
					So(resultPolicySet.Policy(0).HasVerb("POST"), ShouldBeTrue)

					Convey("The delegation is attached to the parent delegation", func() {
						delegated, _ := resultPolicySet.Delegation()
						So(delegated.Policy(0).HasVerb("GET"), ShouldBeTrue)
						So(delegated.Policy(0).HasVerb("POST"), ShouldBeFalse)
					})
				})

			})
		})

	})
}
