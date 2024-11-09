package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	natsproxy "github.com/aliheydarabadii/nats-proxy"
	"github.com/nats-io/nats.go"
)

type Config struct {
	NatsURL string `json:"nats_url"`
	Port    string `json:"port"`
}

func loadConfig(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func main() {
	configFile := flag.String("config", "config.json", "Path to config file")
	natsURL := flag.String("nats-url", nats.DefaultURL, "NATS server URL")
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	config, err := loadConfig(*configFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("Using command line arguments instead")
	} else {
		if config.NatsURL != "" {
			natsURL = &config.NatsURL
		}
		if config.Port != "" {
			port = &config.Port
		}
	}

	proxyConn, err := nats.Connect(*natsURL)
	if err != nil {
		fmt.Printf("Error connecting to NATS: %v\n", err)
		os.Exit(1)
	}

	proxy, err := natsproxy.NewNatsProxy(proxyConn)
	if err != nil {
		fmt.Printf("Error creating proxy: %v\n", err)
		os.Exit(1)
	}

	defer proxyConn.Close()

	listenAddr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Starting server on port %s\n", *port)
	if err := http.ListenAndServe(listenAddr, proxy); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
