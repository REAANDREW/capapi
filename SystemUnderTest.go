package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
)

//SystemUnderTest is a facade to start the APISecurityGateway, APISecurityGatewayProxy and a Fake Upstream HTTP API in order to test the solution.
type SystemUnderTest struct {
	APIGateway             APISecurityGateway
	APIGatewayProxy        *httptest.Server
	APIGatewayControlProxy *httptest.Server
	FakeEndpoint           *httptest.Server
	KeyStore               KeyStore
	ResponseBody           string
	ResponseCode           int
	ServerListener         net.Listener
}

//CreateSystemUnderTest takes a KeyStore and returns a pointer to a new instance of a SystemUnderTest
func CreateSystemUnderTest(keyStore KeyStore) *SystemUnderTest {
	instance := &SystemUnderTest{}

	instance.KeyStore = keyStore

	instance.FakeEndpoint = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//var expectedResponseBody = "You Made It Baby, Yeh!"
		var expectedResponseBody = instance.ResponseBody
		w.WriteHeader(instance.ResponseCode)
		fmt.Fprintln(w, expectedResponseBody)
	}))

	var gatewayProxy = APISecurityGatewayProxy{
		UpStream: ":12345",
	}

	instance.APIGatewayProxy = httptest.NewUnstartedServer(gatewayProxy.Handler())
	instance.APIGatewayControlProxy = httptest.NewUnstartedServer(gatewayProxy.ControlHandler())

	return instance
}

//SetResponseBody sets the response which the fake upstream HTTP API will return.
func (instance *SystemUnderTest) SetResponseBody(value string) {
	instance.ResponseBody = value
}

//SetResponseCode sets the response code which the fake upstream HTTP API will return.
func (instance *SystemUnderTest) SetResponseCode(value int) {
	instance.ResponseCode = value
}

//Start starts all the systems under the System Under Test.
func (instance *SystemUnderTest) Start() {
	instance.FakeEndpoint.Start()
	instance.APIGatewayProxy.Start()
	instance.APIGatewayControlProxy.Start()

	serverListener, err := net.Listen("tcp", ":12345")
	instance.ServerListener = serverListener

	CheckError(err)

	upStreamURL, _ := url.Parse(instance.FakeEndpoint.URL)
	var gateway = APISecurityGateway{
		UpStream: *upStreamURL,
		KeyStore: instance.KeyStore,
	}
	go gateway.Start(serverListener)
}

//Stop stops all the systems under the System Under Test.
func (instance *SystemUnderTest) Stop() {
	instance.ServerListener.Close()
	instance.FakeEndpoint.Close()
	instance.APIGatewayProxy.Close()
	instance.APIGatewayControlProxy.Close()
	time.Sleep(1 * time.Millisecond)
}
