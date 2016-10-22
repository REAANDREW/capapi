package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDelegation(t *testing.T) {

	log.SetLevel(log.ErrorLevel)

	Convey("Delegation", t, func() {
		var keystore = CreateInProcKeyStore()
		var sut = CreateSystemUnderTest(keystore)
		var expectedResponseBody = "You Made It Baby, Yeh!"
		var expectedResponseCode = 200

		sut.SetResponseBody(expectedResponseBody)
		sut.SetResponseCode(expectedResponseCode)
		defer sut.Stop()
		sut.Start()

		delegateURL := sut.APIGatewayControlProxy.URL + "/delegate"

		Convey("unrestricted capability", func() {

			client := &http.Client{}
			key, policyBytes := NewPolicySetBuilder().Build()
			keystore.Set(key, policyBytes)

			Convey("returns a different key", func() {
				var jsonBytes = []byte(`[]`)

				req, _ := http.NewRequest("POST", delegateURL, bytes.NewBuffer(jsonBytes))
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				delegatedKey, _ := ioutil.ReadAll(resp.Body)

				So(delegatedKey, ShouldNotEqual, key)
				So(len(delegatedKey), ShouldEqual, len(key))

				decodedKey, _ := base64.StdEncoding.DecodeString(key)
				So(len(decodedKey), ShouldEqual, KeySizeBytes())

				decodedDelegatedKey, _ := base64.StdEncoding.DecodeString(key)
				So(len(decodedDelegatedKey), ShouldEqual, KeySizeBytes())
			})
		})

		Convey("with initial capability", func() {

			client := &http.Client{}
			key, policyBytes := NewPolicySetBuilder().Build()
			keystore.Set(key, policyBytes)

			Convey("with port policy delegation", func() {
				var jsonBytes = []byte(`[{"verbs":["put"]}]`)

				req, _ := http.NewRequest("POST", delegateURL, bytes.NewBuffer(jsonBytes))
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				delegatedKey, _ := ioutil.ReadAll(resp.Body)

				Convey("must succeed", func() {
					newReq, _ := http.NewRequest("put", sut.APIGatewayProxy.URL+"/something", bytes.NewBuffer(jsonBytes))
					newReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", delegatedKey))
					newResp, err := client.Do(newReq)
					if err != nil {
						panic(err)
					}

					So(newResp.StatusCode, ShouldEqual, 200)
				})

				Convey("must fail", func() {
					newReq, _ := http.NewRequest("get", sut.APIGatewayProxy.URL+"/something", bytes.NewBuffer(jsonBytes))
					newReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", delegatedKey))
					newResp, err := client.Do(newReq)
					if err != nil {
						panic(err)
					}

					So(newResp.StatusCode, ShouldEqual, 401)
				})
			})

			Convey("with path delegation", func() {
				var jsonBytes = []byte(`[{"path":"/some/path"}]`)

				req, _ := http.NewRequest("POST", delegateURL, bytes.NewBuffer(jsonBytes))
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				delegatedKey, _ := ioutil.ReadAll(resp.Body)

				Convey("must succeed", func() {
					newReq, _ := http.NewRequest("get", sut.APIGatewayProxy.URL+"/some/path", bytes.NewBuffer(jsonBytes))
					newReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", delegatedKey))
					newResp, err := client.Do(newReq)
					if err != nil {
						panic(err)
					}

					So(newResp.StatusCode, ShouldEqual, 200)
				})

				Convey("must fail", func() {
					newReq, _ := http.NewRequest("get", sut.APIGatewayProxy.URL+"/something", bytes.NewBuffer(jsonBytes))
					newReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", delegatedKey))
					newResp, err := client.Do(newReq)
					if err != nil {
						panic(err)
					}

					So(newResp.StatusCode, ShouldEqual, 401)
				})
			})

			Convey("with header delegation", func() {
				var jsonBytes = []byte(`[{"headers":{ "A":["1","2","3"] }}]`)

				req, _ := http.NewRequest("POST", delegateURL, bytes.NewBuffer(jsonBytes))
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				delegatedKey, _ := ioutil.ReadAll(resp.Body)

				Convey("must succeed", func() {
					newReq, _ := http.NewRequest("get", sut.APIGatewayProxy.URL, bytes.NewBuffer(jsonBytes))
					newReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", delegatedKey))
					newReq.Header.Set("A", "2")
					newResp, err := client.Do(newReq)
					if err != nil {
						panic(err)
					}

					So(newResp.StatusCode, ShouldEqual, 200)
				})

				Convey("must fail", func() {
					newReq, _ := http.NewRequest("get", sut.APIGatewayProxy.URL, bytes.NewBuffer(jsonBytes))
					newReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", delegatedKey))
					newReq.Header.Set("A", "5")
					newResp, err := client.Do(newReq)
					if err != nil {
						panic(err)
					}

					So(newResp.StatusCode, ShouldEqual, 401)
				})
			})
		})
	})
}
