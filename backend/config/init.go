package config

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
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
	once      sync.Once
	cfg       atomic.Value
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

func loadConfig() Config {
	c := Config{}
	err := c.parseConfig()
	if err != nil {
		log.Fatal(err)
	}
	err = c.initOAuth2Config()
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func (c *Config) parseConfig() error {
	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal([]byte(data), c)
	if err != nil {
		return err
	}
	s, err := os.Stat(cfgPath)
	if err != nil {
		return err
	}
	c.modTime = s.ModTime().UTC()
	return nil
}

func (c *Config) initOAuth2Config() error {
	c.OAuth2Configs = make(map[string]oauth2.Config)
	for _, p := range c.OAuth2Providers {
		endpoint, ok := endpoints[p.Id]
		if !ok {
			return errors.New("Illegal OAuth2 configuration")
		}
		c.OAuth2Configs[p.Id] = oauth2.Config{
			ClientID:     p.ClientId,
			ClientSecret: p.ClientSecret,
			Scopes:       p.Scopes,
			Endpoint:     endpoint,
			RedirectURL:  c.OAuth2RedirectUrl,
		}
	}
	return nil
}

func initConfigFileWatcher() {
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()
		watcher.Add(cfgPath)
		for {
			select {
			case err := <-watcher.Errors:
				log.Println("initConfigFileWatcher:", err)
			case <-watcher.Events:
				log.Println("Config file changed")
				cfg.Store(loadConfig())
			}
		}
	}()
}