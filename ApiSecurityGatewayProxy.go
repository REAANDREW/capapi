package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	capnp "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"

	log "github.com/Sirupsen/logrus"
)

func getProxy(apiKeyValue string, factory HTTPProxyFactoryAPI, ctx context.Context) (HTTPProxyAPI, error) {
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	apiKeyObj, _ := NewAPIKey(seg)
	apiKeyObj.SetValue(apiKeyValue)

	proxyResult, err := factory.GetHTTPProxy(ctx, func(p HTTPProxyFactoryAPI_getHTTPProxy_Params) error {
		return p.SetKey(apiKeyObj)
	}).Struct()

	if err != nil {
		return HTTPProxyAPI{}, err
	}

	proxy := proxyResult.Proxy()
	return proxy, nil
}

//APISecurityGatewayProxy allows a caller to call the APISecurityGateway using the Cap'N Proto procotol.
type APISecurityGatewayProxy struct {
	UpStream string
}

//ControlHandler returns the http.HandlerFunc to allow for Delegations and Revocations to be requested via HTTP
func (instance APISecurityGatewayProxy) ControlHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var policies []PolicyJSONDto

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&policies)

		CheckError(err)

		w.Header().Set("X-CAPAPI", "1")
		apiKeyValue, err := ParseAuthorization(r)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		c, _ := net.Dial("tcp", instance.UpStream)
		defer c.Close()

		conn := rpc.NewConn(rpc.StreamTransport(c), rpc.ConnLog(nil))
		defer conn.Close()

		ctx := context.Background()
		factory := HTTPProxyFactoryAPI{Client: conn.Bootstrap(ctx)}
		proxy, err := getProxy(apiKeyValue, factory, ctx)

		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		result := proxy.Delegate(ctx, func(p HTTPProxyAPI_delegate_Params) error {
			_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
			if err != nil {
				panic(err)
			}
			policySet, err := NewPolicySet(seg)
			CheckError(err)

			policyList, err := NewPolicy_List(seg, int32(len(policies)))
			CheckError(err)

			for index, jsonPolicy := range policies {
				policy, err := NewPolicy(seg)
				CheckError(err)

				verbList, err := capnp.NewTextList(seg, int32(len(jsonPolicy.Verbs)))
				CheckError(err)

				for verbIndex, verb := range jsonPolicy.Verbs {
					verbList.Set(verbIndex, verb)
				}

				policy.SetVerbs(verbList)
				policyList.Set(index, policy)
			}

			policySet.SetPolicies(policyList)

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

		c, _ := net.Dial("tcp", instance.UpStream)
		defer c.Close()

		conn := rpc.NewConn(rpc.StreamTransport(c), rpc.ConnLog(nil))
		defer conn.Close()

		ctx := context.Background()
		factory := HTTPProxyFactoryAPI{Client: conn.Bootstrap(ctx)}
		proxy, err := getProxy(apiKeyValue, factory, ctx)

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

			headerList, err := NewKeyValue_List(seg, int32(len(r.Header)))
			CheckError(err)

			count := 0
			for key, value := range r.Header {
				log.WithFields(log.Fields{
					"key":   key,
					"value": strings.Join(value, ","),
				}).Debug("processing request header")
				header, err := NewKeyValue(seg)
				CheckError(err)
				header.SetKey(key)
				header.SetValue(strings.Join(value, ","))
				headerList.Set(count, header)
				count++
			}

			request.SetHeaders(headerList)

			return p.SetRequestObj(request)
		}).Response()

		response, _ := result.Struct()

		body, _ := response.Body()
		status := response.Status()

		w.WriteHeader(int(status))
		fmt.Fprint(w, body)
	})
}
