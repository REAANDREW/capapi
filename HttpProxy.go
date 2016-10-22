package main

import (
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	capnp "zombiezen.com/go/capnproto2"
)

//HTTPProxy is used to call the upstream HTTP API which is being secured
type HTTPProxy struct {
	APIKey   APIKey
	scope    PolicySet
	upStream url.URL
	keyStore KeyStore
}

//Request transforms the request from the rpc structures into a HTTP Request and sends to the upstream HTTP API which is being secured
func (instance HTTPProxy) Request(call HTTPProxyAPI_request) error {

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

//Delegate allows the capability currently assigned to the HTTPProxy to be delegated
func (instance HTTPProxy) Delegate(call HTTPProxyAPI_delegate) error {
	scope, _ := call.Params.Scope()

	log.WithFields(log.Fields(scope.Map())).Debug("delegating PolicySet")

	//Create a new policy set
	saveMsg, saveSeg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	parentClone := instance.scope.Clone(saveSeg)

	parentClone.SetDelegation(scope)

	saveMsg.SetRootPtr(parentClone.ToPtr())

	newKey, err := CreateKey()
	CheckError(err)

	newScopeBytes, err := saveMsg.Marshal()
	CheckError(err)

	instance.keyStore.Set(newKey, newScopeBytes)

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	key, err := NewAPIKey(seg)
	CheckError(err)

	key.SetValue(newKey)

	return call.Results.SetKey(key)
}

//Revoke allows the current capability
func (instance HTTPProxy) Revoke(call HTTPProxyAPI_revoke) error {
	call.Results.SetResult(false)
	return nil
}
