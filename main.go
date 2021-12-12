package main

import (
	"flag"
	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/services"
	"log"
	"os"
)

func main() {
	flag.Parse()

	err := config.Load()
	if err != nil {
		panic("Failed to load config")
	}

	err = services.Start(config.Cfg.Listen)
	if err != nil {
		log.Printf("Failed to start: %v\n", err)
		os.Exit(1)
	}
}
