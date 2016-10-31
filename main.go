package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app      = kingpin.New("capapi", "Object Capability Based HTTP API Security Gateway")
	debug    = app.Flag("debug", "Enable debug mode.").Bool()
	serverIP = app.Flag("server", "Server address.").Default("127.0.0.1").IP()

	gateway         = app.Command("gateway", "Start a new gateway")
	gatewayHost     = gateway.Flag("host", "The hostname for the gateway").Default("0.0.0.0").String()
	gatewayPort     = gateway.Flag("port", "The port for the gateway").Default("27520").String()
	gatewayUpstream = gateway.Flag("upstream", "The upstream API").Required().String()

	proxy            = app.Command("http-proxy", "Start a new gateway http proxy")
	proxyHost        = proxy.Flag("host", "The hostname for the proxy").Default("0.0.0.0").String()
	proxyPort        = proxy.Flag("port", "The port for the proxy").Default("80").String()
	proxyControlPort = proxy.Flag("control-port", "The port for the proxy control").Default("27526").String()
	upstreamHost     = proxy.Flag("upstream-host", "The hostname for the upstream gateway").Default("0.0.0.0").String()
	upstreamPort     = proxy.Flag("upstream-port", "The port for the upstream gateway").Default("27520").String()
)

func main() {
	log.SetLevel(log.ErrorLevel)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case gateway.FullCommand():

		listenAddress := fmt.Sprintf("%s:%s", *gatewayHost, *gatewayPort)
		fmt.Println(fmt.Sprintf("%s", listenAddress))

		gatewayListener, err := net.Listen("tcp", listenAddress)
		CheckError(err)

		upstreamURL, err := url.Parse(*gatewayUpstream)
		CheckError(err)

		//Initialize the struct with the relevant information including hosts and ports
		var gateway = APISecurityGateway{
			UpStream: *upstreamURL,
			KeyStore: CreateInProcKeyStore(),
		}
		go gateway.Start(gatewayListener)
		fmt.Println(fmt.Sprintf("server running..."))

	case proxy.FullCommand():

		fmt.Println("You are running a proxy")

		gatewayAddress := fmt.Sprintf("%s:%s", *gatewayHost, *gatewayPort)
		var gatewayProxy = APISecurityGatewayProxy{
			UpStream: gatewayAddress,
		}

		s := &http.Server{
			Addr:           ":8080",
			Handler:        gatewayProxy.Handler(),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		go s.ListenAndServe()
		fmt.Println(fmt.Sprintf("http-to-rpc bridge running..."))

	}

	var wg sync.WaitGroup
	wg.Add(1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			fmt.Println("exiting...")
			wg.Done()
		}
	}()
	wg.Wait()
}
