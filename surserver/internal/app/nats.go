package app

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func startEmbeddedNatsServer(appName string, opts NATSConfig) (*natsserver.Server, error) {
	host, port, err := splitHostPort(opts.URL)
	if err != nil {
		return nil, err
	}

	serverOpts := &natsserver.Options{
		ServerName:      fmt.Sprintf("%s-nats-server", appName),
		DontListen:      opts.Private,
		JetStream:       true,
		JetStreamDomain: appName,
		Host:            host,
		Port:            port,
	}

	ns, err := natsserver.NewServer(serverOpts)

	if err != nil {
		return nil, err
	}

	if opts.Logging {
		ns.ConfigureLogger()
	}

	ns.Start()

	if !ns.ReadyForConnections(5 * time.Second) {
		return nil, nats.ErrTimeout
	}

	return ns, nil
}

func splitHostPort(url string) (string, int, error) {
	address := strings.Split(url, "//")[1]
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}

	return host, port, nil
}

func connectToEmbeddedNATS(appName string, ns *natsserver.Server, opts NATSConfig) (*nats.Conn, error) {
	clientOpts := []nats.Option{
		nats.Name(fmt.Sprintf("%s-nats-client", appName)),
	}
	if opts.Private {
		clientOpts = append(clientOpts, nats.InProcessServer(ns))
	}
	nc, err := nats.Connect(opts.URL, clientOpts...)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

func connectToExternalNATS(opts NATSConfig) (*nats.Conn, error) {
	nc, err := nats.Connect(opts.URL)
	if err != nil {
		return nil, err
	}

	return nc, nil
}
