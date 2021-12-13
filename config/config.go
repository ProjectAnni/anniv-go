package config

import (
	"flag"
	"github.com/go-yaml/yaml"
	"os"
)

type Config struct {
	SiteName    string            `yaml:"site_name"`
	Description string            `yaml:"description"`
	Listen      string            `yaml:"listen"`
	DBPath      string            `yaml:"db_path"`
	Enforce2FA  bool              `yaml:"enforce_2fa"`
	Headers     map[string]string `yaml:"headers"`
}

var Cfg = Config{
	SiteName:    "Anniv",
	Description: "",
	Listen:      ":8080",
	DBPath:      "data.db",
	Enforce2FA:  false,
}

var path = flag.String("path", "config.yaml", "")

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
	return yaml.NewEncoder(f).Encode(&Cfg)
}
