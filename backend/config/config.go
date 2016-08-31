package config

import (
	"log"
	"time"

	"golang.org/x/oauth2"
)

type Config struct {
	ListenAddr        string                   `yaml:"listen-addr"`
	TLSCertFile       string                   `yaml:"tls-cert-file"`
	TLSKeyFile        string                   `yaml:"tls-key-file"`
	OAuth2RedirectUrl string                   `yaml:"oauth2-redirect-url"`
	OAuth2State       OAuth2State              `yaml:"oauth2-state"`
	OAuth2Providers   []OAuth2Provider         `yaml:"oauth2-providers"`
	OAuth2Configs     map[string]oauth2.Config `yaml:"-"`
	modTime           time.Time                `yaml:"-"`
}

type OAuth2State struct {
	TokenIssuer  string        `yaml:"token-issuer`
	TokenSignKey []byte        `yaml:"token-sign-key`
	Expiration   time.Duration `yaml:"expiration`
}

type OAuth2Provider struct {
	Id           string   `json:"id" yaml:"id"`
	Name         string   `json:"name" yaml:"name"`
	ClientId     string   `json:"-" yaml:"client-id"`
	ClientSecret string   `json:"-" yaml:"client-secret"`
	PublicKey    string   `json:"-" yaml:"public-key"`
	Scopes       []string `json:"-" yaml:"scopes"`
	IconURL      string   `json:"iconUrl" yaml:"icon"`
}

func GetConfig() Config {
	once.Do(func() {
		log.Printf("Program started with config file '%s'", cfgPath)
		cfg.Store(loadConfig())
		initConfigFileWatcher()
	})
	return cfg.Load().(Config)
}

func ModTime() time.Time {
	return GetConfig().modTime
}
