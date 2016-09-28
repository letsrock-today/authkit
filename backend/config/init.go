package config

import (
	"encoding/hex"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/bitbucket"
	"golang.org/x/oauth2/clientcredentials"
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
	c.CSRFSecret, err = hex.DecodeString(c.CSRFSecretHex)
	if err != nil {
		return err
	}
	c.OAuth2State.TokenSignKey, err = hex.DecodeString(c.OAuth2State.TokenSignKeyHex)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) initOAuth2Config() error {
	h := c.HydraOAuth2Provider
	c.HydraClientCredentials = clientcredentials.Config{
		ClientID:     h.ClientId,
		ClientSecret: h.ClientSecret,
		Scopes:       h.Scopes,
		TokenURL:     strings.Replace(h.TokenURL, "{base-url}", c.HydraAddr, -1),
	}
	//TODO: Probably, we should use client with less priveleges for oauth2 config below (without access to hydra keys etc.)
	c.HydraOAuth2Config = oauth2.Config{
		ClientID:     h.ClientId,
		ClientSecret: h.ClientSecret,
		Scopes:       h.Scopes,
		Endpoint: oauth2.Endpoint{
			TokenURL: strings.Replace(h.TokenURL, "{base-url}", c.ExternalBaseURL, -1),
			AuthURL:  strings.Replace(h.AuthURL, "{base-url}", c.ExternalBaseURL, -1),
		},
		RedirectURL: c.OAuth2RedirectURL,
	}
	c.HydraOAuth2ConfigInt = oauth2.Config{
		ClientID:     h.ClientId,
		ClientSecret: h.ClientSecret,
		Scopes:       h.Scopes,
		Endpoint: oauth2.Endpoint{
			TokenURL: strings.Replace(h.TokenURL, "{base-url}", c.HydraAddr, -1),
			AuthURL:  strings.Replace(h.AuthURL, "{base-url}", c.HydraAddr, -1),
		},
		RedirectURL: c.OAuth2RedirectURL,
	}
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
			RedirectURL:  c.OAuth2RedirectURL,
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
			// Currently, when config changed with vim, I receive several
			// notifications, and sometimes file doesn't exists at the moment
			// of notification (vim deletes original file and replaces it with
			// tmp copy). We should check event type and reload config only on final
			// event. Probably, short 1-2 second timeout would suffice to debounce
			// events (to simplify logic, or else we would need to check every
			// possible scenario). This is out of scope of this example project.
			case <-watcher.Events:
				log.Println("Config file changed, consider to restart server")
				cfg.Store(loadConfig())
				// This is not enaugh to dynamically load config.
				// We should also send a notification about config change to
				// restart server and routes, but it is out of scope of this
				// example project.
			}
		}
	}()
}
