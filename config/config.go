package config

import (
	"github.com/go-yaml/yaml"
	uuid "github.com/satori/go.uuid"
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
	RequireInvite  bool              `yaml:"require_invite"`
	InviteCode     string            `yaml:"invite_code"`
	AnnilToken     AnnilToken        `yaml:"annil_token"`
}

type AnnilToken struct {
	Enabled    bool   `yaml:"enabled"`
	URL        string `yaml:"url"`
	Secret     string `yaml:"secret"`
	AllowShare bool   `yaml:"allowShare"`
}

var Cfg = Config{
	SiteName:       "Anniv",
	Description:    "",
	Listen:         ":8080",
	Enforce2FA:     false,
	TrustedProxies: []string{"127.0.0.1/32"},
	RepoDir:        "./data/meta",
	RepoURL:        "https://github.com/ProjectAnni/repo.git",
	RequireInvite:  false,
	InviteCode:     uuid.NewV4().String(),
	AnnilToken: AnnilToken{
		Enabled:    false,
		URL:        "",
		Secret:     "",
		AllowShare: false,
	},
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
