package config

import (
	"flag"
	"io/ioutil"
	"log"

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

var (
	cfgPath   string
	cfg       = Config{}
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

func initOAuth2Config() {
	cfg.OAuth2Configs = make(map[string]oauth2.Config)
	for _, p := range cfg.OAuth2Providers {
		endpoint, ok := endpoints[p.Id]
		if !ok {
			log.Fatal("Illegal OAuth2 configuration")
		}
		cfg.OAuth2Configs[p.Id] = oauth2.Config{
			ClientID:     p.ClientId,
			ClientSecret: p.ClientSecret,
			Scopes:       p.Scopes,
			Endpoint:     endpoint,
			RedirectURL:  cfg.OAuth2RedirectUrl,
		}
	}
}
