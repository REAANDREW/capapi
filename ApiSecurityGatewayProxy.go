package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

				policy.SetPath(jsonPolicy.Path)

				verbList, err := capnp.NewTextList(seg, int32(len(jsonPolicy.Verbs)))
				CheckError(err)

				for verbIndex, verb := range jsonPolicy.Verbs {
					verbList.Set(verbIndex, verb)
				}

				policy.SetVerbs(verbList)

				headerKeyValueList, err := NewKeyValuePolicy_List(seg, int32(len(jsonPolicy.Headers)))
				CheckError(err)

				headerCount := 0
				for headerKey, headerValues := range jsonPolicy.Headers {
					keyValuePolicy, err := NewKeyValuePolicy(seg)
					CheckError(err)

					keyValuePolicy.SetKey(headerKey)

					headerValueList, err := capnp.NewTextList(seg, int32(len(headerValues)))
					CheckError(err)

					for index, headerValue := range headerValues {
						log.WithFields(log.Fields{
							"headerKey":   headerKey,
							"headerValue": headerValue,
						}).Debug("adding headers to policy")
						headerValueList.Set(index, headerValue)
					}

					keyValuePolicy.SetValues(headerValueList)
					headerKeyValueList.Set(headerCount, keyValuePolicy)
					headerCount++
				}
				policy.SetHeaders(headerKeyValueList)

				queryKeyValueList, err := NewKeyValuePolicy_List(seg, int32(len(jsonPolicy.Query)))
				CheckError(err)

				queryCount := 0

				log.WithFields(log.Fields{
					"numberOfFields": len(jsonPolicy.Query),
				}).Debug("About to iterate over Query")

				for queryKey, queryValues := range jsonPolicy.Query {
					keyValuePolicy, err := NewKeyValuePolicy(seg)
					CheckError(err)

					keyValuePolicy.SetKey(queryKey)

					queryValueList, err := capnp.NewTextList(seg, int32(len(queryValues)))
					CheckError(err)

					log.WithFields(log.Fields{
						"queryKey": queryKey,
					}).Debug("query key")

					for index, queryValue := range queryValues {
						log.WithFields(log.Fields{
							"queryValue": queryValue,
						}).Debug("query value for key")
						queryValueList.Set(index, queryValue)
					}

					keyValuePolicy.SetValues(queryValueList)
					queryKeyValueList.Set(queryCount, keyValuePolicy)
					queryCount++
				}
				policy.SetQuery(queryKeyValueList)

				policyList.Set(index, policy)
			}

			policySet.SetPolicies(policyList)

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
		log.WithFields(log.Fields{
			"url": r.URL.String(),
		}).Debug("Received request")

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

			queryList, err := NewKeyValue_List(seg, int32(len(r.URL.Query())))
			CheckError(err)

			count = 0
			for key, value := range r.URL.Query() {
				log.WithFields(log.Fields{
					"key":   key,
					"value": strings.Join(value, ","),
				}).Debug("processing request query")
				query, err := NewKeyValue(seg)
				CheckError(err)
				query.SetKey(key)
				query.SetValue(strings.Join(value, ","))
				queryList.Set(count, query)
				count++
			}

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
