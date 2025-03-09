package main

import (
	"flag"
	"log"
	"os"

	"github.com/lstep/surroundhome/surserver/internal/app"
	"github.com/lstep/surroundhome/surserver/internal/mods/rest-nats"
)

func main() {
	os.Exit(Run(os.Args[1:]))
}

func Run(args []string) int {
	// Define flags
	flagSet := flag.NewFlagSet("surserver", flag.ExitOnError)
	configPath := flagSet.String("c", "config.yaml", "Path to configuration file")

	// Parse flags
	if err := flagSet.Parse(args); err != nil {
		log.Printf("Failed to parse flags: %v", err)
		return 1
	}

	config, err := app.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	myApp := app.New(*config)

	// Add modules
	myApp.AddModule(&rest.RestModule{})

	if err := myApp.Start(); err != nil {
		log.Fatalf("Failed to start app: %v", err)
	}

	// Wait for app to stop
	<-myApp.StopApp

	if err := myApp.Stop(); err != nil {
		log.Printf("Failed to stop app: %v", err)
	}

	return 0
}
