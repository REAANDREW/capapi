package tests

import (
	"fmt"
	"github.com/reaandrew/capapi/core"
	"github.com/reaandrew/capapi/gateway"
	"github.com/reaandrew/capapi/proxy"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
)

type SystemUnderTest struct {
	APIGateway      gateway.ApiSecurityGateway
	APIGatewayProxy *httptest.Server
	FakeEndpoint    *httptest.Server
	KeyStore        core.KeyStore
	ResponseBody    string
	ResponseCode    int
	ServerListener  net.Listener
}

func CreateSystemUnderTest(keyStore core.KeyStore) *SystemUnderTest {
	instance := &SystemUnderTest{}

	instance.KeyStore = keyStore

	instance.FakeEndpoint = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//var expectedResponseBody = "You Made It Baby, Yeh!"
		var expectedResponseBody = instance.ResponseBody
		w.WriteHeader(instance.ResponseCode)
		fmt.Fprintln(w, expectedResponseBody)
	}))

	var gatewayProxy = proxy.ApiSecurityGatewayProxy{
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

	core.CheckError(err)

	upStreamURL, _ := url.Parse(instance.FakeEndpoint.URL)
	var gateway = gateway.ApiSecurityGateway{
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
