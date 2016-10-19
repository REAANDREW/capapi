package gateway

import (
	"net/url"

	capnp "zombiezen.com/go/capnproto2"

	"github.com/reaandrew/capapi/capability"
	"github.com/reaandrew/capapi/core"
)

type HTTPProxyFactory struct {
	KeyStore core.KeyStore
	UpStream url.URL
}

func (instance HTTPProxyFactory) GetHTTPProxy(call capability.HTTPProxyFactoryAPI_getHTTPProxy) error {
	apiKey, _ := call.Params.Key()
	apiKeyValue, _ := apiKey.Value()

	bytesValue, err := instance.KeyStore.Get(apiKeyValue)
	if err != nil {
		return err
		//panic(err)
	}

	msg, _ := capnp.Unmarshal(bytesValue)
	scope, _ := capability.ReadRootPolicySet(msg)

	server := capability.HTTPProxyAPI_ServerToClient(HttpProxy{
		APIKey:   apiKey,
		scope:    scope,
		upStream: instance.UpStream,
	})

	return call.Results.SetProxy(server)
}
