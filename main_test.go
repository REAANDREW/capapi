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
)

/*

Due to composition it simply means that given an API Key, when it is delegated, the parent scope is always evaluated first
therefor when the new scope is evaluated it must be further defined than the parent otherwise it would not get evaluated

WIN, WIN, WIN, WIN!!

*/

func CreateKeyStore() KeyStore {
	keyStore := InProcessKeyStore{
		Keys: map[string][]byte{},
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
		var keystore = CreateKeyStore()
		var sut = CreateSystemUnderTest(keystore)
		var expectedResponseBody = "You Made It Baby, Yeh!"
		var expectedResponseCode = 200

		sut.SetResponseBody(expectedResponseBody)
		sut.SetResponseCode(expectedResponseCode)
		defer sut.Stop()
		sut.Start()

		Convey("with unrestricted access", func() {
			key, bytes := NewPolicySetBuilder().Build()
			keystore.Set(key, bytes)

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

			key, bytes := NewPolicySetBuilder().
				WithPolicy(NewPolicyBuilder().WithVerb("PUT")).
				Build()

			keystore.Set(key, bytes)

			Convey("must succeed", func() {
				client := &http.Client{}

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

		Convey("with exact path policy", func() {
			okPath := "/some/path"

			key, bytes := NewPolicySetBuilder().
				WithPolicy(NewPolicyBuilder().WithPath(okPath)).
				Build()

			keystore.Set(key, bytes)

			Convey("must succeed", func() {
				client := &http.Client{}

				exactPath := sut.APIGatewayProxy.URL + okPath

				req, _ := http.NewRequest("PUT", exactPath, nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
				resp, err := client.Do(req)
				if err != nil {
					log.Error(err)
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					panic(err)
				}

				So(resp.StatusCode, ShouldEqual, expectedResponseCode)
				So(strings.Trim(string(body), "\n"), ShouldEqual, expectedResponseBody)
			})

			Convey("must fail", func() {
				client := &http.Client{}

				exactPath := sut.APIGatewayProxy.URL + "/someother/path"
				req, _ := http.NewRequest("POST", exactPath, nil)

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
