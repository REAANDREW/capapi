package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app      = kingpin.New("capapi", "Object Capability Based HTTP API Security Gateway")
	debug    = app.Flag("debug", "Enable debug mode.").Bool()
	serverIP = app.Flag("server", "Server address.").Default("127.0.0.1").IP()

	gateway         = app.Command("gateway", "Start a new gateway")
	gatewayUpstream = gateway.Flag("upstream", "The upstream API").Required().String()

	proxy        = app.Command("http-proxy", "Start a new gateway http proxy")
	proxyHost    = proxy.Flag("proxy-host", "The hostname for the proxy").Strings()
	proxyPort    = proxy.Flag("proxy-port", "The port for the proxy").Strings()
	upstreamHost = proxy.Flag("upstream-host", "The hostname for the upstream gateway").Strings()
	upstreamPort = proxy.Flag("upstream-port", "The port for the upstream gateway").Strings()
)

func main() {
	log.SetLevel(log.ErrorLevel)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case gateway.FullCommand():
		fmt.Println("You are running a server")

	case proxy.FullCommand():
		fmt.Println("You are running a proxy")
	}
}
