package config

import (
	"os"

	"github.com/go-yaml/yaml"
	"github.com/google/uuid"
)

type Config struct {
	SiteName       string            `yaml:"site_name"`
	Description    string            `yaml:"description"`
	Listen         string            `yaml:"listen"`
	Enforce2FA     bool              `yaml:"enforce_2fa"`
	Headers        map[string]string `yaml:"headers"`
	TrustedProxies []string          `yaml:"trusted_proxies"`
	RepoURL        string            `yaml:"repo_url"`
	RequireInvite  bool              `yaml:"require_invite"`
	InviteCode     string            `yaml:"invite_code"`
	AnnilToken     []AnnilToken      `yaml:"annil_token"`
	Debug          DebugConfig       `yaml:"debug"`
	EnableMeta     bool              `yaml:"enable_meta"`
}

type AnnilToken struct {
	Enabled      bool   `yaml:"enabled"`
	Name         string `yaml:"name"`
	URL          string `yaml:"url"`
	AdminBaseURL string `yaml:"admin_base_url,omitempty"`
	Credential   string `yaml:"credential"`
	AllowShare   bool   `yaml:"allow_share"`
}

type DebugConfig struct {
	Enabled        bool   `yaml:"enabled"`
	MemProfilePath string `yaml:"mem_profile_path"`
}

var Cfg = Config{
	SiteName:       "Anniv",
	Description:    "",
	Listen:         ":8080",
	Enforce2FA:     false,
	TrustedProxies: []string{"127.0.0.1/32"},
	RepoURL:        "https://github.com/ProjectAnni/repo.git",
	RequireInvite:  false,
	InviteCode:     uuid.NewString(),
	AnnilToken: []AnnilToken{
		{
			Enabled:    false,
			Name:       "Default Library",
			URL:        "",
			Credential: "",
			AllowShare: false,
		},
	},
	Debug: DebugConfig{
		Enabled:        false,
		MemProfilePath: "mem.prof",
	},
	EnableMeta: true,
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
