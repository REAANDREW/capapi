package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	capnp "zombiezen.com/go/capnproto2"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/gorilla/mux"
)

/*

Due to composition it simply means that given an API Key, when it is delegated, the parent scope is always evaluated first
therefor when the new scope is evaluated it must be further defined than the parent otherwise it would not get evaluated

WIN, WIN, WIN, WIN!!

*/

const apiPort = 50000
const rpcPort = 60000

var rpcAddr = fmt.Sprintf(":%d", rpcPort)

func CreateKeyStore() keyStore {
	var key = "unsecure_key_number_1"
	msg, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	scope, _ := NewRootHTTPProxyScope(seg)
	scope.SetPath("Bang Bang")
	textList, _ := capnp.NewTextList(seg, 1)
	textList.Set(0, "GET")
	scope.SetVerbs(textList)

	byteValue, _ := msg.Marshal()

	keyStore := inProcessKeyStore{
		keys: map[string][]byte{
			key: byteValue,
		},
	}

	return keyStore
}

func StartAPISecurityGateway(keyStore keyStore) {
	serverListener, _ := net.Listen("tcp", rpcAddr)
	upStreamURL, _ := url.Parse(fmt.Sprintf("http://localhost:%d", apiPort))
	var gateway = apiSecurityGateway{
		upStream: *upStreamURL,
		keyStore: keyStore,
	}
	go gateway.start(serverListener)
}

func CreateAPISecurityGatewayProxy() *httptest.Server {
	var gatewayProxy = apiSecurityGatewayProxy{
		upStream: rpcAddr,
	}

	ts := httptest.NewUnstartedServer(gatewayProxy.handler())
	return ts
}

func CreateFakeEndpoint() *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, "You Made It!")
	}))
	return server
}

func TestSomething(t *testing.T) {
	Convey("Does something", t, func() {

		Convey("Test my knowledge of MUX", func() {
			req, _ := http.NewRequest("GET", "http://localhost:3000/fubar/2", nil)
			r := mux.NewRouter()
			r.Path("/fubar/{id:(1|2)}")
			var match mux.RouteMatch
			result := r.Match(req, &match)

			So(result, ShouldEqual, true)
		})

		Convey("Request", func() {
			keyStore := CreateKeyStore()
			StartAPISecurityGateway(keyStore)

			gatewayProxy := CreateAPISecurityGatewayProxy()
			defer gatewayProxy.Close()
			gatewayProxy.Start()

			fakeEndpoint := CreateFakeEndpoint()
			defer fakeEndpoint.Close()
			fakeEndpoint.Start()

			//Set a key without restriction
			Convey("without restriction", func() {

				client := &http.Client{}
				req, _ := http.NewRequest("GET", fakeEndpoint.URL, nil)
				req.Header.Add("If-None-Match", `W/"wyzzy"`)
				resp, _ := client.Do(req)

				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				So(string(body), ShouldEqual, "You Made It!\n")
			})
		})

		Convey("Delegation", func() {
			Convey("From an ALL powerful master", func() {
				/*
					json := `{
						"paths" : ["/fubar/:id"]
						"pathValues" : {
							"id" : "(1|2)"
						}
					}`

					client := &http.Client{}

					req, err := http.NewRequest("GET", "http://example.com", nil)
					req.Header.Add("If-None-Match", `W/"wyzzy"`)
					resp, err := client.Do(req)
				*/
			})
		})
	})
}
