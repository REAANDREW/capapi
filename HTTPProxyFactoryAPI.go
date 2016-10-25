package main

import (
	"context"
	"net"

	capnp "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

//EXTENSION METHODS FOR THE GENERATED CODE!

//NewHTTPProxyFactoryAPI encapsulates the details for creating a new HTTPProxyFactoryAPI
func NewHTTPProxyFactoryAPI(ctx context.Context, address string) HTTPProxyFactoryAPI {
	c, _ := net.Dial("tcp", address)
	conn := rpc.NewConn(rpc.StreamTransport(c), rpc.ConnLog(nil))
	factory := HTTPProxyFactoryAPI{Client: conn.Bootstrap(ctx)}
	return factory
}

//GetProxy gets the reference to the proxy which is assigned to the supplied apiKey
func (instance HTTPProxyFactoryAPI) GetProxy(apiKey string, ctx context.Context) (HTTPProxyAPI, error) {
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	apiKeyObj, _ := NewAPIKey(seg)
	apiKeyObj.SetValue(apiKey)

	proxyResult, err := instance.GetHTTPProxy(ctx, func(p HTTPProxyFactoryAPI_getHTTPProxy_Params) error {
		return p.SetKey(apiKeyObj)
	}).Struct()

	if err != nil {
		return HTTPProxyAPI{}, err
	}

	proxy := proxyResult.Proxy()
	return proxy, nil
}
