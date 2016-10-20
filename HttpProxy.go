package main

import (
	"io/ioutil"
	"net/http"
	"net/url"

	capnp "zombiezen.com/go/capnproto2"
)

type HttpProxy struct {
	APIKey   APIKey
	scope    PolicySet
	upStream url.URL
}

func (instance HttpProxy) Request(call HTTPProxyAPI_request) error {

	req, _ := call.Params.RequestObj()

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	response, _ := NewHTTPResponse(seg)

	if !instance.scope.Validate(req) {
		response.SetStatus(401)
		return call.Results.SetResponse(response)
	}

	client := &http.Client{}
	upstreamRequest, _ := http.NewRequest("GET", instance.upStream.String(), nil)
	upstreamResponse, _ := client.Do(upstreamRequest)
	defer upstreamResponse.Body.Close()
	body, _ := ioutil.ReadAll(upstreamResponse.Body)

	response.SetBody(string(body))
	response.SetStatus(uint32(upstreamResponse.StatusCode))

	return call.Results.SetResponse(response)
}

func (instance HttpProxy) Delegate(call HTTPProxyAPI_delegate) error {
	//scope, _ := call.Params.Scope()

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	key, _ := NewAPIKey(seg)
	key.SetValue("You Key, Sir!")
	return call.Results.SetKey(key)
}

func (instance HttpProxy) Revoke(call HTTPProxyAPI_revoke) error {
	call.Results.SetResult(false)
	return nil
}
