package config

import (
	"encoding/hex"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/linkedin"

	"github.com/pkg/errors"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/peculiarproviders/deezer"
)

const (
	defPath = "./env"
	defName = "dev"
)

var (
	endpoints = map[string]oauth2.Endpoint{
		// Add as much as you need,
		// ID here should correspond
		// to ID in tne configuration file.
		"fb":       facebook.Endpoint,
		"google":   google.Endpoint,
		"linkedin": linkedin.Endpoint,
		"deezer":   deezer.Endpoint,
	}
)

type Config struct {
	PrivateProviderID               string            `mapstructure:"-"`
	PrivateProviderIDTrustedContext string            `mapstructure:"-"`
	ListenAddr                      string            `mapstructure:"listen-addr"`
	TLSCertFile                     string            `mapstructure:"tls-cert-file"`
	TLSKeyFile                      string            `mapstructure:"tls-key-file"`
	TLSInsecureSkipVerify           bool              `mapstructure:"tls-insecure-skip-verify"`
	HydraAddr                       string            `mapstructure:"hydra-addr"`
	ExternalBaseURL                 string            `mapstructure:"external-base-url"`
	OAuth2RedirectURL               string            `mapstructure:"oauth2-redirect-url"`
	ChallengeLifespan               time.Duration     `mapstructure:"challenge-lifespan"`
	ConfirmationLinkLifespan        time.Duration     `mapstructure:"confirmation-link-lifespan"`
	AuthCookieName                  string            `mapstructure:"auth-cookie-name"`
	AuthHeaderName                  string            `mapstructure:"auth-header-name"`
	EmailConfig                     EmailConfig       `mapstructure:"email-config"`
	MongoDB                         DBConfig          `mapstructure:"mongodb"`
	OAuth2State                     OAuth2State       `mapstructure:"oauth2-state"`
	PrivateOAuth2Provider           *OAuth2Provider   `mapstructure:"private-oauth2-provider"`
	OAuth2Providers                 []*OAuth2Provider `mapstructure:"oauth2-providers"`
	OAuth2ClientCredentials         *clientcredentials.Config
	ModTime                         time.Time
}

func (c *Config) init() {
	c.PrivateProviderID = "authkit"
	c.PrivateProviderIDTrustedContext = "authkit-trusted"
	c.PrivateOAuth2Provider.ID = c.PrivateProviderID

	c.OAuth2State.init()

	c.OAuth2ClientCredentials = &clientcredentials.Config{
		ClientID:     c.PrivateOAuth2Provider.ClientID,
		ClientSecret: c.PrivateOAuth2Provider.ClientSecret,
		Scopes:       c.PrivateOAuth2Provider.Scopes,
		TokenURL: strings.Replace(
			c.PrivateOAuth2Provider.TokenURL,
			"{base-url}",
			c.HydraAddr,
			-1),
	}

	p := c.PrivateOAuth2Provider
	p.OAuth2Config = c.newOAuth2Config(p, false)
	p.PrivateOAuth2Config = c.newOAuth2Config(p, true)

	for _, p := range c.OAuth2Providers {
		p.OAuth2Config = c.newOAuth2Config(p, false)
	}

	c.ModTime = time.Now()
}

func (c *Config) newOAuth2Config(
	p *OAuth2Provider,
	privateConfig bool) authkit.OAuth2Config {
	var endpoint oauth2.Endpoint
	if p.ID == c.PrivateProviderID {
		baseURL := c.ExternalBaseURL
		if privateConfig {
			baseURL = c.HydraAddr
		}
		tokenURL := strings.Replace(p.TokenURL, "{base-url}", baseURL, -1)
		authURL := strings.Replace(p.AuthURL, "{base-url}", baseURL, -1)
		endpoint = oauth2.Endpoint{
			TokenURL: tokenURL,
			AuthURL:  authURL,
		}
	} else {
		var ok bool
		endpoint, ok = endpoints[p.ID]
		if !ok || privateConfig {
			panic(errors.Errorf(
				"Illegal OAuth2 configuration, for: %s, %t",
				p.ID,
				privateConfig))
		}
	}
	cfg := &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Scopes:       p.Scopes,
		Endpoint:     endpoint,
		RedirectURL:  c.OAuth2RedirectURL,
	}
	// for peculiar providers we need custom wrappers
	if p.ID == "deezer" {
		return &deezer.Config{cfg}
	}
	return cfg
}

func (c Config) ToAuthkitType() authkit.Config {
	app := []authkit.OAuth2Provider{}
	for _, p := range c.OAuth2Providers {
		app = append(app, p.ToAuthkitType())
	}
	ac := authkit.Config{
		OAuth2Providers:       app,
		PrivateOAuth2Provider: c.PrivateOAuth2Provider.ToAuthkitType(),
		OAuth2State:           c.OAuth2State.ToAuthkitType(),
		AuthCookieName:        c.AuthCookieName,
		ModTime:               c.ModTime,
	}
	return ac
}

type EmailConfig struct {
	Sender     string `mapstructure:"sender"`
	SenderPass string `mapstructure:"sender-pass"`
	MailServer string `mapstructure:"server"`
	MailPort   string `mapstructure:"port"`
}

type DBConfig struct {
	URL  string `mapstructure:"url"`
	Name string `mapstructure:"name"`
}

type OAuth2State struct {
	TokenIssuer     string        `mapstructure:"token-issuer"`
	TokenSignKeyHex string        `mapstructure:"token-sign-key"`
	Expiration      time.Duration `mapstructure:"expiration"`
	TokenSignKey    []byte
}

func (s *OAuth2State) init() {
	var err error
	s.TokenSignKey, err = hex.DecodeString(s.TokenSignKeyHex)
	if err != nil {
		panic(err)
	}
}

func (s OAuth2State) ToAuthkitType() authkit.OAuth2State {
	as := authkit.OAuth2State{
		TokenIssuer:  s.TokenIssuer,
		Expiration:   s.Expiration,
		TokenSignKey: s.TokenSignKey,
	}
	return as
}

type OAuth2Provider struct {
	ID                  string   `mapstructure:"id"`
	Name                string   `mapstructure:"name"`
	ClientID            string   `mapstructure:"client-id"`
	ClientSecret        string   `mapstructure:"client-secret"`
	PublicKey           string   `mapstructure:"public-key"`
	Scopes              []string `mapstructure:"scopes"`
	IconURL             string   `mapstructure:"icon"`
	TokenURL            string   `mapstructure:"token-url"`
	AuthURL             string   `mapstructure:"auth-url"`
	OAuth2Config        authkit.OAuth2Config
	PrivateOAuth2Config authkit.OAuth2Config
}

func (p OAuth2Provider) ToAuthkitType() authkit.OAuth2Provider {
	ap := authkit.OAuth2Provider{
		ID:                  p.ID,
		Name:                p.Name,
		IconURL:             p.IconURL,
		OAuth2Config:        p.OAuth2Config,
		PrivateOAuth2Config: p.PrivateOAuth2Config,
	}
	return ap
}
