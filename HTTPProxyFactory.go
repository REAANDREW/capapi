package main

import (
	"net/url"

	capnp "zombiezen.com/go/capnproto2"
)

//HTTPProxyFactory return a HTTPProxy to the calling rpc client which has the relevant capability attached.
type HTTPProxyFactory struct {
	KeyStore KeyStore
	UpStream url.URL
}

//GetHTTPProxy the capability is what determines when the incoming request can be made or not.
func (instance HTTPProxyFactory) GetHTTPProxy(call HTTPProxyFactoryAPI_getHTTPProxy) error {
	apiKey, _ := call.Params.Key()
	apiKeyValue, _ := apiKey.Value()

	bytesValue, err := instance.KeyStore.Get(apiKeyValue)
	if err != nil {
		return err
		//panic(err)
	}

	msg, _ := capnp.Unmarshal(bytesValue)
	scope, _ := ReadRootPolicySet(msg)

	server := HTTPProxyAPI_ServerToClient(HTTPProxy{
		APIKey:   apiKey,
		scope:    scope,
		upStream: instance.UpStream,
		keyStore: instance.KeyStore,
	})

	return call.Results.SetProxy(server)
}
