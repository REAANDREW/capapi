package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	//	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	capnp "zombiezen.com/go/capnproto2"
	//	"github.com/gorilla/mux"
)

/*

Due to composition it simply means that given an API Key, when it is delegated, the parent scope is always evaluated first
therefor when the new scope is evaluated it must be further defined than the parent otherwise it would not get evaluated

WIN, WIN, WIN, WIN!!

*/

const key = "unsecure_key_number_1"

func CreateUnrestrictedKeyStore() keyStore {
	msg, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	scope, _ := NewRootHTTPProxyScope(seg)
	textList, _ := capnp.NewTextList(seg, 0)
	scope.SetVerbs(textList)

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

var proxyFactory = httpProxyFactory{
	keyStore: CreateUnrestrictedKeyStore(),
}

func TestProcess(t *testing.T) {
	Convey("Returns", t, func() {
		Convey("Successfully", func() {
			Convey("with unrestricted access", func() {
				var expectedResponseBody = "You Made It Baby, Yeh!"

				fakeEndpoint := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					fmt.Fprintln(w, expectedResponseBody)
				}))

				defer fakeEndpoint.Close()
				fakeEndpoint.Start()

				serverListener, _ := net.Listen("tcp", ":12345")
				upStreamURL, _ := url.Parse(fakeEndpoint.URL)
				var gateway = apiSecurityGateway{
					upStream: *upStreamURL,
					keyStore: CreateUnrestrictedKeyStore(),
				}
				go gateway.start(serverListener)

				var gatewayProxy = apiSecurityGatewayProxy{
					upStream: ":12345",
				}

				ts := httptest.NewUnstartedServer(gatewayProxy.handler())
				defer ts.Close()
				ts.Start()

				client := &http.Client{}
				req, _ := http.NewRequest("GET", ts.URL, nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
				resp, _ := client.Do(req)
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)

				So(resp.StatusCode, ShouldEqual, 200)
				So(strings.Trim(string(body), "\n"), ShouldEqual, expectedResponseBody)
			})
		})
	})

}
