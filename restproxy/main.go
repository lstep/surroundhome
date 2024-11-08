package main

import (
	"net/http"

	natsproxy "github.com/aliheydarabadii/nats-proxy"
	"github.com/nats-io/nats.go"
)

func main() {
	proxyConn, _ := nats.Connect(nats.DefaultURL)
	//proxyConn, _ := nats.Connect("http://ailocal:4333")
	proxy, _ := natsproxy.NewNatsProxy(proxyConn)
	defer proxyConn.Close()
	http.ListenAndServe(":8080", proxy)
}
