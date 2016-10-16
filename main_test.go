package main

import (
	"context"
	"fmt"
	//	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	capnp "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
	//. "github.com/smartystreets/goconvey/convey"
	//	"github.com/gorilla/mux"
)

/*

Due to composition it simply means that given an API Key, when it is delegated, the parent scope is always evaluated first
therefor when the new scope is evaluated it must be further defined than the parent otherwise it would not get evaluated

WIN, WIN, WIN, WIN!!

*/

const apiPort = 50000
const rpcPort = 60000
const key = "unsecure_key_number_1"

var rpcAddr = fmt.Sprintf(":%d", rpcPort)

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

func CreateRestrictedKeyStore() keyStore {
	var key = "unsecure_key_number_1"
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

func StartAPISecurityGateway(keyStore keyStore) {
	serverListener, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		panic(err)
	}
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

var proxyFactory = httpProxyFactory{
	keyStore: CreateRestrictedKeyStore(),
}

func handleServer(c net.Conn) error {
	main := HTTPProxyFactory_ServerToClient(proxyFactory)
	conn := rpc.NewConn(rpc.StreamTransport(c), rpc.MainInterface(main.Client))
	err := conn.Wait()
	return err
}

func handleClient(ctx context.Context, c net.Conn) error {
	conn := rpc.NewConn(rpc.StreamTransport(c))
	defer conn.Close()

	factory := HTTPProxyFactory{Client: conn.Bootstrap(ctx)}

	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))

	apiKey, _ := NewAPIKey(seg)
	apiKey.SetValue(key)
	proxy := factory.GetHTTPProxy(ctx, func(p HTTPProxyFactory_getHTTPProxy_Params) error {
		return p.SetKey(apiKey)
	}).Proxy()

	result := proxy.Request(ctx, func(p HTTPProxy_request_Params) error {
		_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
		request, _ := NewHTTPRequest(seg)
		request.SetPath("/fubar")
		request.SetVerb("GET")
		return p.SetRequestObj(request)
	}).Response()

	response, err := result.Struct()

	if err != nil {
		panic(err)
	}

	fmt.Println(response)

	return err
}

func TestDoSomething(t *testing.T) {
	/*
		c1, _ := net.Listen("tcp", ":12345")
		defer c1.Close()

		c2, _ := net.Dial("tcp", ":12345")
		defer c2.Close()

		go func() {
			for {
				if conn, err := c1.Accept(); err == nil {
					//If err is nil then that means that data is available for us so we take up this data and pass it to a new goroutine
					t.Log("Go it!!")
					go handleServer(conn)
				} else {
					continue
				}
			}
		}()
	*/
	serverListener, err := net.Listen("tcp", ":12345")
	if err != nil {
		panic(err)
	}

	fakeEndpoint := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, "You Made It Baby, Yeh!")
	}))

	defer fakeEndpoint.Close()
	fakeEndpoint.Start()

	upStreamURL, _ := url.Parse(fakeEndpoint.URL)
	var gateway = apiSecurityGateway{
		upStream: *upStreamURL,
		keyStore: CreateRestrictedKeyStore(),
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

	fmt.Println(fmt.Sprintf("Response code %v", resp.StatusCode))

	//	handleClient(context.Background(), c2)

}

/*

	fakeEndpoint := CreateFakeEndpoint()
	defer fakeEndpoint.Close()
	fakeEndpoint.Start()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", gatewayProxy.URL, nil)
	req.Header.Set("Authorization", "Bearer something")
	resp, _ := client.Do(req)

	//So(resp.Header.Get("X-CAPAPI"), ShouldEqual, "1")
	//So(resp.StatusCode, ShouldEqual, http.StatusUnauthorized)
	fmt.Println(fmt.Sprintf("Status Code: %d", resp.StatusCode))

*/
