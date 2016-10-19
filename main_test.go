package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	capnp "zombiezen.com/go/capnproto2"
)

/*

Due to composition it simply means that given an API Key, when it is delegated, the parent scope is always evaluated first
therefor when the new scope is evaluated it must be further defined than the parent otherwise it would not get evaluated

WIN, WIN, WIN, WIN!!

*/

const key = "unsecure_key_number_1"

func CreateKeyStore() keyStore {
	msg, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	policySet, _ := NewRootPolicySet(seg)

	policyList, _ := NewPolicy_List(seg, 1)

	policy, _ := NewPolicy(seg)

	textList, _ := capnp.NewTextList(seg, 0)

	policy.SetVerbs(textList)

	policyList.Set(0, policy)

	policySet.SetPolicies(policyList)

	byteValue, _ := msg.Marshal()

	keyStore := inProcessKeyStore{
		keys: map[string][]byte{
			key: byteValue,
		},
	}

	return keyStore
}

func CreateFakeEndpoint() *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, "You Made It!")
	}))
	return server
}

func TestCapapi(t *testing.T) {

	log.SetLevel(log.ErrorLevel)

	Convey("API Call", t, func() {
		Convey("with unrestricted access", func() {
			var keystore = CreateKeyStore()
			var sut = CreateSystemUnderTest(keystore)
			var expectedResponseBody = "You Made It Baby, Yeh!"
			var expectedResponseCode = 200

			sut.setResponseBody(expectedResponseBody)
			sut.setResponseCode(expectedResponseCode)
			defer sut.stop()
			sut.start()
			client := &http.Client{}
			req, _ := http.NewRequest("GET", sut.APIGatewayProxy.URL, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
			resp, _ := client.Do(req)
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)

			So(resp.StatusCode, ShouldEqual, expectedResponseCode)
			So(strings.Trim(string(body), "\n"), ShouldEqual, expectedResponseBody)
		})

		Convey("with port policy", func() {
			var keystore = CreateKeyStore()
			var sut = CreateSystemUnderTest(keystore)
			var expectedResponseBody = "You Made It Baby, Yeh!"
			var expectedResponseCode = 200

			sut.setResponseBody(expectedResponseBody)
			sut.setResponseCode(expectedResponseCode)
			defer sut.stop()
			sut.start()

			key, bytes := newPolicySetBuilder().
				withPolicy(newPolicyBuilder().withVerb("PUT")).
				build()

			keystore.Set(key, bytes)

			Convey("must succeed", func() {
				client := &http.Client{}
				log.WithFields(log.Fields{
					"url": sut.APIGatewayProxy.URL,
				}).Info("API Gateway Proxy URL")
				req, _ := http.NewRequest("PUT", sut.APIGatewayProxy.URL, nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
				resp, err := client.Do(req)
				if err != nil {
					log.Error(err)
				}
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)

				So(resp.StatusCode, ShouldEqual, expectedResponseCode)
				So(strings.Trim(string(body), "\n"), ShouldEqual, expectedResponseBody)
			})

			Convey("must fail", func() {
				client := &http.Client{}
				log.WithFields(log.Fields{
					"url": sut.APIGatewayProxy.URL,
				}).Info("API Gateway Proxy URL")
				req, _ := http.NewRequest("POST", sut.APIGatewayProxy.URL, nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
				resp, err := client.Do(req)
				if err != nil {
					log.Error(err)
				}
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)

				So(resp.StatusCode, ShouldEqual, 401)
				So(strings.Trim(string(body), "\n"), ShouldEqual, "")
			})
		})
	})
}
