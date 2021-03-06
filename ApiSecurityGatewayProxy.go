package main

import (
	"context"
	"fmt"
	"net/http"

	capnp "zombiezen.com/go/capnproto2"

	log "github.com/Sirupsen/logrus"
)

//APISecurityGatewayProxy allows a caller to call the APISecurityGateway using the Cap'N Proto procotol.
type APISecurityGatewayProxy struct {
	UpStream string
}

//ControlHandler returns the http.HandlerFunc to allow for Delegations and Revocations to be requested via HTTP
func (instance APISecurityGatewayProxy) ControlHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CAPAPI", "1")

		var policies = DecodePolicyJSONDtos(r.Body)

		apiKeyValue, err := ParseAuthorization(r)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.Background()

		factory := NewHTTPProxyFactoryAPI(ctx, instance.UpStream)

		proxy, err := factory.GetProxy(apiKeyValue, ctx)

		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		result := proxy.Delegate(ctx, func(p HTTPProxyAPI_delegate_Params) error {
			policySet, err := NewPolicySetFromPolicyJSONDtos(policies)
			CheckError(err)
			log.WithFields(log.Fields(policySet.Map())).Debug("sending delegation")
			return p.SetScope(policySet)
		}).Key()

		key, _ := result.Struct()
		w.WriteHeader(int(201))

		keyValue, err := key.Value()
		CheckError(err)

		fmt.Fprint(w, keyValue)

	})
}

//Handler returns the http.HandlerFunc which handles the request via http and proxies it to the rpc server
func (instance APISecurityGatewayProxy) Handler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CAPAPI", "1")
		apiKeyValue, err := ParseAuthorization(r)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.Background()
		factory := NewHTTPProxyFactoryAPI(ctx, instance.UpStream)
		proxy, err := factory.GetProxy(apiKeyValue, ctx)

		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		result := proxy.Request(ctx, func(p HTTPProxyAPI_request_Params) error {
			_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
			if err != nil {
				panic(err)
			}
			request, _ := NewHTTPRequest(seg)
			request.SetVerb(r.Method)
			request.SetPath(r.URL.Path)

			headerList, err := KeyValueListFromMap(r.Header)
			CheckError(err)
			request.SetHeaders(headerList)

			queryList, err := KeyValueListFromMap(r.URL.Query())
			CheckError(err)
			request.SetQuery(queryList)

			return p.SetRequestObj(request)
		}).Response()

		response, _ := result.Struct()

		body, _ := response.Body()
		status := response.Status()

		w.WriteHeader(int(status))
		fmt.Fprint(w, body)
	})
}
