package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

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
		apiKey := r.Header.Get("Authorization")
		apiKeySplit := strings.Split(apiKey, ":")
		apiKeyValue := strings.Trim(apiKeySplit[1], " ")

		c, _ := net.Dial("tcp", instance.upStream)
		defer c.Close()

		conn := rpc.NewConn(rpc.StreamTransport(c))
		conn.Close()

		ctx := context.Background()
		factory := HTTPProxyFactory{Client: conn.Bootstrap(ctx)}

		_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

		apiKeyObj, _ := NewAPIKey(seg)
		apiKeyObj.SetValue(apiKeyValue)
		proxy := factory.GetHTTPProxy(ctx, func(p HTTPProxyFactory_getHTTPProxy_Params) error {
			return p.SetKey(apiKeyObj)
		}).Proxy()

		result := proxy.Request(ctx, func(p HTTPProxy_request_Params) error {
			_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
			request, _ := NewHTTPRequest(seg)
			request.SetPath("/fubar")
			request.SetVerb("PATCH")
			return p.SetRequestObj(request)
		}).Response()

		response, _ := result.Struct()
		body, _ := response.Body()
		status := response.Status()

		w.WriteHeader(int(status))
		fmt.Fprintln(w, body)
	})
}

type apiSecurityGateway struct {
	upStream url.URL
	keyStore keyStore
	listener net.Listener
}

func (instance apiSecurityGateway) start(listener net.Listener) {
	instance.listener = listener

	for {
		if conn, err := instance.listener.Accept(); err == nil {
			go func() {
				main := HTTPProxyFactory_ServerToClient(httpProxyFactory{})
				conn := rpc.NewConn(rpc.StreamTransport(conn), rpc.MainInterface(main.Client))
				err := conn.Wait()
				panic(err)
			}()
		} else {
			continue
		}
	}
}
