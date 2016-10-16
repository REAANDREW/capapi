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

type SystemUnderTest struct {
	APIGateway      apiSecurityGateway
	APIGatewayProxy *httptest.Server
	FakeEndpoint    *httptest.Server
	ResponseBody    string
	ResponseCode    int
}

func CreateSystemUnderTest() *SystemUnderTest {
	instance := &SystemUnderTest{}

	instance.FakeEndpoint = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//var expectedResponseBody = "You Made It Baby, Yeh!"
		var expectedResponseBody = instance.ResponseBody
		w.WriteHeader(instance.ResponseCode)
		fmt.Fprintln(w, expectedResponseBody)
	}))

	var gatewayProxy = apiSecurityGatewayProxy{
		upStream: ":12345",
	}

	instance.APIGatewayProxy = httptest.NewUnstartedServer(gatewayProxy.handler())

	return instance
}
func (instance *SystemUnderTest) setResponseBody(value string) {
	instance.ResponseBody = value
}

func (instance *SystemUnderTest) setResponseCode(value int) {
	instance.ResponseCode = value
}

func (instance *SystemUnderTest) start() {
	instance.FakeEndpoint.Start()
	instance.APIGatewayProxy.Start()

	serverListener, _ := net.Listen("tcp", ":12345")
	upStreamURL, _ := url.Parse(instance.FakeEndpoint.URL)
	var gateway = apiSecurityGateway{
		upStream: *upStreamURL,
		keyStore: CreateUnrestrictedKeyStore(),
	}
	go gateway.start(serverListener)
}

func (instance *SystemUnderTest) stop() {
	instance.FakeEndpoint.Close()
	instance.APIGatewayProxy.Close()
}

var proxyFactory = httpProxyFactory{
	keyStore: CreateUnrestrictedKeyStore(),
}

func TestProcess(t *testing.T) {
	Convey("With", t, func() {
		Convey("unrestricted access", func() {
			Convey("it returns successfully", func() {
				var expectedResponseBody = "You Made It Baby, Yeh!"
				var expectedResponseCode = 200

				var sut = CreateSystemUnderTest()
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
		})
		Convey("restricted access to single verb", func() {
			Convey("it returns successfully", func() {

			})
		})
	})

}
