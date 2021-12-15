package config

import (
	"flag"
	"github.com/go-yaml/yaml"
	"os"
)

type Config struct {
	SiteName       string            `yaml:"site_name"`
	Description    string            `yaml:"description"`
	Listen         string            `yaml:"listen"`
	Enforce2FA     bool              `yaml:"enforce_2fa"`
	Headers        map[string]string `yaml:"headers"`
	TrustedProxies []string          `yaml:"trusted_proxies"`
}

var Cfg = Config{
	SiteName:       "Anniv",
	Description:    "",
	Listen:         ":8080",
	Enforce2FA:     false,
	TrustedProxies: []string{"127.0.0.1/32"},
}

var path = flag.String("conf", "config.yaml", "")

func Load() error {
	f, err := os.Open(*path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := Save(); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(&Cfg)
}

func Save() error {
	f, err := os.Create(*path)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(&Cfg)
}
