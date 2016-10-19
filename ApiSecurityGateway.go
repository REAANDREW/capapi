package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"

	capnp "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
)

//LOOKING TO MOVE TO https://github.com/hashicorp/yamux
//LOOKS REALLY USEFUL ESPECIALLY TURNING AROUND THE STREAMS

type apiSecurityGatewayProxy struct {
	upStream string
}

func (instance apiSecurityGatewayProxy) handler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CAPAPI", "1")
		apiKeyValue, err := parseAuthorization(r)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		c, _ := net.Dial("tcp", instance.upStream)
		defer c.Close()

		conn := rpc.NewConn(rpc.StreamTransport(c))

		ctx := context.Background()
		factory := HTTPProxyFactory{Client: conn.Bootstrap(ctx)}

		_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

		apiKeyObj, _ := NewAPIKey(seg)
		apiKeyObj.SetValue(apiKeyValue)

		proxyResult, err := factory.GetHTTPProxy(ctx, func(p HTTPProxyFactory_getHTTPProxy_Params) error {
			return p.SetKey(apiKeyObj)
		}).Struct()

		if err != nil {
			log.Error(err)
			c.Close()
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		proxy := proxyResult.Proxy()

		result := proxy.Request(ctx, func(p HTTPProxy_request_Params) error {
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

type apiSecurityGateway struct {
	upStream url.URL
	keyStore keyStore
}

func (instance apiSecurityGateway) start(listener net.Listener) {
	for {
		if c, err := listener.Accept(); err == nil {
			go func() {
				main := HTTPProxyFactory_ServerToClient(httpProxyFactory{
					keyStore: instance.keyStore,
					upStream: instance.upStream,
				})
				conn := rpc.NewConn(rpc.StreamTransport(c), rpc.MainInterface(main.Client))
				err := conn.Wait()
				if err != nil && err != io.EOF {
					log.Error(err)
				}
			}()
		} else {
			continue
		}
	}
}
