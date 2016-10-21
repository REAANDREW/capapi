package main

import (
	"io"
	"net"
	"net/url"

	log "github.com/Sirupsen/logrus"

	"zombiezen.com/go/capnproto2/rpc"
)

//LOOKING TO MOVE TO https://github.com/hashicorp/yamux
//LOOKS REALLY USEFUL ESPECIALLY TURNING AROUND THE STREAMS

//APISecurityGateway is the rpc server which return an instance of a
//HTTPProxyFactory for the caller to use.
type APISecurityGateway struct {
	UpStream url.URL
	KeyStore KeyStore
}

//Start will start the APISecurityGateway.
//It takes a stream based listener.
func (instance APISecurityGateway) Start(listener net.Listener) {
	for {
		if c, err := listener.Accept(); err == nil {
			go func() {
				main := HTTPProxyFactoryAPI_ServerToClient(HTTPProxyFactory{
					KeyStore: instance.KeyStore,
					UpStream: instance.UpStream,
				})
				conn := rpc.NewConn(rpc.StreamTransport(c), rpc.MainInterface(main.Client), rpc.ConnLog(nil))

				err := conn.Wait()
				if err != nil && err != io.EOF && err.Error() != "rpc: aborted by remote: rpc: shutdown" {
					log.Error(err)
				}
			}()
		} else {
			continue
		}
	}
}
