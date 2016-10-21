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

func CreateFakeEndpoint() *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, "You Made It!")
	}))
	return server
}

func TestCapapi(t *testing.T) {

	log.SetLevel(log.ErrorLevel)

	var keystore = CreateInProcKeyStore()
	var sut = CreateSystemUnderTest(keystore)
	var expectedResponseBody = "You Made It Baby, Yeh!"
	var expectedResponseCode = 200

	sut.SetResponseBody(expectedResponseBody)
	sut.SetResponseCode(expectedResponseCode)
	defer sut.Stop()
	sut.Start()

	Convey("with unrestricted access", t, func() {
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

	Convey("with port policy", t, func() {
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

	Convey("with exact path policy", t, func() {
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

	Convey("with templated path policy", t, func() {
		key, bytes := NewPolicySetBuilder().
			WithPolicy(NewPolicyBuilder().WithPath("/clients/{clientId:(1|2)}/data")).
			Build()

		keystore.Set(key, bytes)

		Convey("must succeed", func() {
			client := &http.Client{}

			exactPath := sut.APIGatewayProxy.URL + "/clients/1/data"

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

			exactPath := sut.APIGatewayProxy.URL + "/clients/3/data"

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

	Convey("with a header policy", t, func() {
		key, bytes := NewPolicySetBuilder().
			WithPolicy(NewPolicyBuilder().WithHeader("X-Something", []string{"1"})).
			Build()

		keystore.Set(key, bytes)

		Convey("must succeed", func() {
			client := &http.Client{}

			req, _ := http.NewRequest("PUT", sut.APIGatewayProxy.URL, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
			req.Header.Set("X-Something", "1")

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

			req, _ := http.NewRequest("PUT", sut.APIGatewayProxy.URL, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
			req.Header.Set("X-Something", "2")

			resp, err := client.Do(req)
			if err != nil {
				log.Error(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			So(resp.StatusCode, ShouldEqual, 401)
			So(strings.Trim(string(body), "\n"), ShouldEqual, "")
		})
	})

	Convey("with a querystring policy", t, func() {
		key, bytes := NewPolicySetBuilder().
			WithPolicy(NewPolicyBuilder().WithQuery("a", []string{})).
			Build()

		keystore.Set(key, bytes)

		Convey("must succeed", func() {
			client := &http.Client{}

			req, _ := http.NewRequest("PUT", sut.APIGatewayProxy.URL, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
			req.URL.Query().Add("a", "1")

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

			req, _ := http.NewRequest("PUT", sut.APIGatewayProxy.URL, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
			req.URL.Query().Add("b", "1")

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

		Convey("must fail with no query string", func() {
			client := &http.Client{}

			req, _ := http.NewRequest("PUT", sut.APIGatewayProxy.URL, nil)
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
	})
}
