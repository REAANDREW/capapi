package main

import (
	"io/ioutil"
	"net/http"
	"net/url"

	capnp "zombiezen.com/go/capnproto2"
)

var caps = map[string][]byte{}

type verbValidator struct {
	Scope HTTPProxyScope
}

func (instance HTTPProxyScope) validate(request HTTPRequest) bool {
	verbs, _ := instance.Verbs()

	if verbs.Len() == 0 {
		return true
	}

	verb, _ := request.Verb()
	for i := 0; i < verbs.Len(); i++ {
		scopedVerb, _ := verbs.At(i)
		if verb == scopedVerb {
			return true
		}
	}

	return false
}

type httpProxy struct {
	APIKey   APIKey
	scope    HTTPProxyScope
	upStream url.URL
}

func (instance httpProxy) Request(call HTTPProxy_request) error {

	req, _ := call.Params.RequestObj()

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	response, _ := NewHTTPResponse(seg)

	if !instance.scope.validate(req) {
		response.SetStatus(401)
		return call.Results.SetResponse(response)
	}

	//path, _ := req.Path()
	//verb, _ := req.Verb()

	client := &http.Client{}
	upstreamRequest, _ := http.NewRequest("GET", instance.upStream.String(), nil)
	upstreamResponse, _ := client.Do(upstreamRequest)
	defer upstreamResponse.Body.Close()
	body, _ := ioutil.ReadAll(upstreamResponse.Body)

	response.SetBody(string(body))
	response.SetStatus(uint32(upstreamResponse.StatusCode))

	return call.Results.SetResponse(response)
}

func (instance httpProxy) Delegate(call HTTPProxy_delegate) error {
	//scope, _ := call.Params.Scope()

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	key, _ := NewAPIKey(seg)
	key.SetValue("You Key, Sir!")
	return call.Results.SetKey(key)
}

func (instance httpProxy) Revoke(call HTTPProxy_revoke) error {
	call.Results.SetResult(false)
	return nil
}

type httpProxyFactory struct {
	keyStore keyStore
	upStream url.URL
}

func (instance httpProxyFactory) GetHTTPProxy(call HTTPProxyFactory_getHTTPProxy) error {
	apiKey, _ := call.Params.Key()
	apiKeyValue, _ := apiKey.Value()

	bytesValue, err := instance.keyStore.Get(apiKeyValue)
	if err != nil {
		return err
		//panic(err)
	}

	msg, _ := capnp.Unmarshal(bytesValue)
	scope, _ := ReadRootHTTPProxyScope(msg)

	server := HTTPProxy_ServerToClient(httpProxy{
		APIKey:   apiKey,
		scope:    scope,
		upStream: instance.upStream,
	})

	return call.Results.SetProxy(server)
}
