package config

import (
	"flag"
	"io/ioutil"
	"log"
	"time"

	backend_oauth2 "github.com/letsrock-today/hydra-sample/backend/oauth2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/bitbucket"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/heroku"
	"golang.org/x/oauth2/linkedin"
	"golang.org/x/oauth2/odnoklassniki"
	"golang.org/x/oauth2/paypal"
	"golang.org/x/oauth2/slack"
	"golang.org/x/oauth2/uber"
	"golang.org/x/oauth2/vk"
	"gopkg.in/yaml.v2"
)

type AuthConfig struct {
	TokenIssuer           string        `yaml:"token-issuer`
	TokenSignKey          string        `yaml:"token-sign-key`
	OAuth2StateExpiration time.Duration `yaml:"oauth2-state-expiration`
	OAuth2RedirectUrl     string        `yaml:"oauth2-redirect-url"`
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

type Config struct {
	initialized     bool
	OAuth2Providers []OAuth2Provider                       `yaml:"oauth2-providers"`
	AuthConfig      AuthConfig                             `yaml:"auth-configuration"`
	OAuth2Configs   map[string]backend_oauth2.OAuth2Config `yaml:"-"`
}

var (
	cfgPath string
	cfg     = Config{}
	//oauth2config = map[string]oauth2.OAuth2Config{}
	endpoints = map[string]oauth2.Endpoint{
		"fb":        facebook.Endpoint,
		"google":    google.Endpoint,
		"linkedin":  linkedin.Endpoint,
		"ok":        odnoklassniki.Endpoint,
		"paypal":    paypal.Endpoint,
		"vk":        vk.Endpoint,
		"bitbucket": bitbucket.Endpoint,
		"github":    github.Endpoint,
		"heroku":    heroku.Endpoint,
		"slack":     slack.Endpoint,
		"uber":      uber.Endpoint,
	}
)

func init() {
	flag.StringVar(&cfgPath, "config", "./env/dev.yaml", "application's configuration file")
}

func parseConfig() {
	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func initOAuth2Config() error {
	cfg.OAuth2Configs = make(map[string]backend_oauth2.OAuth2Config)
	for _, p := range cfg.OAuth2Providers {
		endpoint, ok := endpoints[p.Id]
		if !ok {
			log.Fatal("Illegal OAuth2 configuration")
		}
		conf := oauth2.Config{
			ClientID:     p.ClientId,
			ClientSecret: p.ClientSecret,
			Scopes:       p.Scopes,
			Endpoint:     endpoint,
			RedirectURL:  cfg.AuthConfig.OAuth2RedirectUrl,
		}
		cfg.OAuth2Configs[p.Id] = &backend_oauth2.Config{Config: &conf}
		//oauth2config[p.Id] = &backend_oauth2.Config{Config: &conf}
	}
	return nil
}

func GetConfig() Config {
	if cfg.initialized {
		return cfg
	}
	log.Printf("Program started with config file '%s'", cfgPath)
	parseConfig()
	initOAuth2Config()
	cfg.initialized = true
	log.Printf("Parsed configuration: %#v", cfg)
	return cfg
}
