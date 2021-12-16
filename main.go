package main

import (
	"flag"
	"fmt"
	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/ProjectAnni/anniv-go/services"
	"log"
	"os"
)

func main() {
	flag.Parse()

	err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	err = meta.Init(config.Cfg.RepoDir, config.Cfg.RepoURL)
	if err != nil {
		log.Printf("Failed to init meta repo: %v\n", err)
		os.Exit(1)
	}

	err = services.Start(config.Cfg.Listen)
	if err != nil {
		log.Printf("Failed to start: %v\n", err)
		os.Exit(1)
	}
}
