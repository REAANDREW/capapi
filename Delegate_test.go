package main

import (
	"bytes"
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

		Convey("unrestricted capability", func() {

			client := &http.Client{}
			key, policyBytes := NewPolicySetBuilder().Build()
			keystore.Set(key, policyBytes)

			Convey("with port policy", func() {

				delegateURL := sut.APIGatewayControlProxy.URL + "/delegate"
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

			})

		})
	})
}
