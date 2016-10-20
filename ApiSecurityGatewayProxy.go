package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	capnp "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"

	log "github.com/Sirupsen/logrus"
)

type ApiSecurityGatewayProxy struct {
	UpStream string
}

func (instance ApiSecurityGatewayProxy) Handler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CAPAPI", "1")
		apiKeyValue, err := ParseAuthorization(r)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		c, _ := net.Dial("tcp", instance.UpStream)
		defer c.Close()

		conn := rpc.NewConn(rpc.StreamTransport(c))

		ctx := context.Background()
		factory := HTTPProxyFactoryAPI{Client: conn.Bootstrap(ctx)}

		_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

		apiKeyObj, _ := NewAPIKey(seg)
		apiKeyObj.SetValue(apiKeyValue)

		proxyResult, err := factory.GetHTTPProxy(ctx, func(p HTTPProxyFactoryAPI_getHTTPProxy_Params) error {
			return p.SetKey(apiKeyObj)
		}).Struct()

		if err != nil {
			log.Error(err)
			c.Close()
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		proxy := proxyResult.Proxy()

		result := proxy.Request(ctx, func(p HTTPProxyAPI_request_Params) error {
			_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
			if err != nil {
				panic(err)
			}
			request, _ := NewHTTPRequest(seg)
			request.SetVerb(r.Method)
			return p.SetRequestObj(request)
		}).Response()

		response, _ := result.Struct()

		body, _ := response.Body()
		status := response.Status()

		w.WriteHeader(int(status))
		fmt.Fprint(w, body)
	})
}
