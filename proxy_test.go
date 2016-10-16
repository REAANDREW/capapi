package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

func testServer(c net.Conn) error {
	main := HTTPProxyFactory_ServerToClient(httpProxyFactory{})
	conn := rpc.NewConn(rpc.StreamTransport(c), rpc.MainInterface(main.Client))
	err := conn.Wait()
	return err
}

func testClient(ctx context.Context, c net.Conn) error {
	conn := rpc.NewConn(rpc.StreamTransport(c))
	defer conn.Close()

	factory := HTTPProxyFactory{Client: conn.Bootstrap(ctx)}

	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))

	apiKey, _ := NewAPIKey(seg)
	apiKey.SetValue("skeleton")
	proxy := factory.GetHTTPProxy(ctx, func(p HTTPProxyFactory_getHTTPProxy_Params) error {
		return p.SetKey(apiKey)
	}).Proxy()

	result := proxy.Request(ctx, func(p HTTPProxy_request_Params) error {
		_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
		request, _ := NewHTTPRequest(seg)
		request.SetPath("/fubar")
		request.SetVerb("PATCH")
		return p.SetRequestObj(request)
	}).Response()

	response, err := result.Struct()

	if err != nil {
		panic(err)
	}

	fmt.Println(response)

	return err
}

func TestProxySecurityGateay(t *testing.T) {
	msg, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	testScope, _ := NewRootHTTPProxyScope(seg)
	testScope.SetPath("Bang Bang")
	textList, _ := capnp.NewTextList(seg, 2)
	textList.Set(0, "GET")
	textList.Set(1, "POST")
	testScope.SetVerbs(textList)

	byteValue, _ := msg.Marshal()
	caps["skeleton"] = byteValue

	c1, serverErr := net.Listen("tcp", ":1234")
	if serverErr != nil {
		log.Fatal("listen error:", serverErr)
	}

	c2, clientErr := net.Dial("tcp", ":1234")
	if clientErr != nil {
		log.Fatal("listen error:", clientErr)
	}

	go func() {
		for {
			if conn, err := c1.Accept(); err == nil {
				//If err is nil then that means that data is available for us so we take up this data and pass it to a new goroutine
				t.Log("Go it!!")
				go testServer(conn)
			} else {
				continue
			}
		}
	}()
	testClient(context.Background(), c2)
}
