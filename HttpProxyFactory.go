package main

import (
	"net/url"

	capnp "zombiezen.com/go/capnproto2"
)

type HTTPProxyFactory struct {
	KeyStore KeyStore
	UpStream url.URL
}

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

	server := HTTPProxyAPI_ServerToClient(HttpProxy{
		APIKey:   apiKey,
		scope:    scope,
		upStream: instance.UpStream,
		keyStore: instance.KeyStore,
	})

	return call.Results.SetProxy(server)
}
