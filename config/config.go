package config

import (
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
	RepoDir        string            `yaml:"repo_dir"`
	RepoURL        string            `yaml:"repo_url"`
}

var Cfg = Config{
	SiteName:       "Anniv",
	Description:    "",
	Listen:         ":8080",
	Enforce2FA:     false,
	TrustedProxies: []string{"127.0.0.1/32"},
	RepoDir:        "./data/meta",
	RepoURL:        "https://github.com/ProjectAnni/repo.git",
}

func Load() error {
	f, err := os.Open(os.Getenv("CONF"))
	if err != nil {
		if os.IsNotExist(err) {
			if err := Save(); err != nil {
				return err
			}
			f, err = os.Open(os.Getenv("CONF"))
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer f.Close()
	err = yaml.NewDecoder(f).Decode(&Cfg)
	if err == nil {
		_ = Save()
	}
	return err
}

func Save() error {
	f, err := os.Create(os.Getenv("CONF"))
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(&Cfg)
}
