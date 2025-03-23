package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/ProjectAnni/anniv-go/services"
)

func main() {
	flag.Parse()
	err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if config.Cfg.Debug.Enabled {
		go func() {
			for {
				f, err := os.Create(config.Cfg.Debug.MemProfilePath)
				if err != nil {
					log.Printf("Failed to create mem profile file: %v\n", err)
					continue
				}

				runtime.GC()
				if err := pprof.WriteHeapProfile(f); err != nil {
					log.Printf("Failed to create heap profile: %v\n", err)
					continue
				}

				f.Close()

				log.Printf("Heap profile saved at %s\n", config.Cfg.Debug.MemProfilePath)

				time.Sleep(time.Minute)
			}
		}()
	}

	if config.Cfg.EnableMeta {
		err = meta.Init("./tmp/meta", config.Cfg.RepoURL)
		if err != nil {
			log.Printf("Failed to init meta repo: %v\n", err)
			os.Exit(1)
		}
	}

	err = services.Start(config.Cfg.Listen)
	if err != nil {
		log.Printf("Failed to start: %v\n", err)
		os.Exit(1)
	}
}
