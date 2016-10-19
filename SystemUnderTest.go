package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
)

type SystemUnderTest struct {
	APIGateway      apiSecurityGateway
	APIGatewayProxy *httptest.Server
	FakeEndpoint    *httptest.Server
	KeyStore        keyStore
	ResponseBody    string
	ResponseCode    int
	ServerListener  net.Listener
}

func CreateSystemUnderTest(keyStore keyStore) *SystemUnderTest {
	instance := &SystemUnderTest{}

	instance.KeyStore = keyStore

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

	serverListener, err := net.Listen("tcp", ":12345")
	instance.ServerListener = serverListener

	checkError(err)

	upStreamURL, _ := url.Parse(instance.FakeEndpoint.URL)
	var gateway = apiSecurityGateway{
		upStream: *upStreamURL,
		keyStore: instance.KeyStore,
	}
	go gateway.start(serverListener)
}

func (instance *SystemUnderTest) stop() {
	instance.FakeEndpoint.Close()
	instance.APIGatewayProxy.Close()
	instance.ServerListener.Close()
	time.Sleep(1 * time.Millisecond)
}
