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
	APIGateway      ApiSecurityGateway
	APIGatewayProxy *httptest.Server
	FakeEndpoint    *httptest.Server
	KeyStore        KeyStore
	ResponseBody    string
	ResponseCode    int
	ServerListener  net.Listener
}

func CreateSystemUnderTest(keyStore KeyStore) *SystemUnderTest {
	instance := &SystemUnderTest{}

	instance.KeyStore = keyStore

	instance.FakeEndpoint = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//var expectedResponseBody = "You Made It Baby, Yeh!"
		var expectedResponseBody = instance.ResponseBody
		w.WriteHeader(instance.ResponseCode)
		fmt.Fprintln(w, expectedResponseBody)
	}))

	var gatewayProxy = ApiSecurityGatewayProxy{
		UpStream: ":12345",
	}

	instance.APIGatewayProxy = httptest.NewUnstartedServer(gatewayProxy.Handler())

	return instance
}
func (instance *SystemUnderTest) SetResponseBody(value string) {
	instance.ResponseBody = value
}

func (instance *SystemUnderTest) SetResponseCode(value int) {
	instance.ResponseCode = value
}

func (instance *SystemUnderTest) Start() {
	instance.FakeEndpoint.Start()
	instance.APIGatewayProxy.Start()

	serverListener, err := net.Listen("tcp", ":12345")
	instance.ServerListener = serverListener

	CheckError(err)

	upStreamURL, _ := url.Parse(instance.FakeEndpoint.URL)
	var gateway = ApiSecurityGateway{
		UpStream: *upStreamURL,
		KeyStore: instance.KeyStore,
	}
	go gateway.Start(serverListener)
}

func (instance *SystemUnderTest) Stop() {
	instance.FakeEndpoint.Close()
	instance.APIGatewayProxy.Close()
	instance.ServerListener.Close()
	time.Sleep(1 * time.Millisecond)
}
