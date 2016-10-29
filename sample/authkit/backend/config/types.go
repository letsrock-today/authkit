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

	"github.com/letsrock-today/hydra-sample/authkit"
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
	}
)

type config struct {
	PrivateProviderID               string                    `mapstructure:"-"`
	PrivateProviderIDTrustedContext string                    `mapstructure:"-"`
	ListenAddr                      string                    `mapstructure:"listen-addr"`
	TLSCertFile                     string                    `mapstructure:"tls-cert-file"`
	TLSKeyFile                      string                    `mapstructure:"tls-key-file"`
	TLSInsecureSkipVerify           bool                      `mapstructure:"tls-insecure-skip-verify"`
	HydraAddr                       string                    `mapstructure:"hydra-addr"`
	ExternalBaseURL                 string                    `mapstructure:"external-base-url"`
	OAuth2RedirectURL               string                    `mapstructure:"oauth2-redirect-url"`
	ChallengeLifespan               time.Duration             `mapstructure:"challenge-lifespan"`
	ConfirmationLinkLifespan        time.Duration             `mapstructure:"confirmation-link-lifespan"`
	AuthCookieName                  string                    `mapstructure:"auth-cookie-name"`
	EmailConfig                     EmailConfig               `mapstructure:"email-config"`
	MongoDB                         DBConfig                  `mapstructure:"mongodb"`
	OAuth2State                     oauth2State               `mapstructure:"oauth2-state" wrapstruct:"-"`
	PrivateOAuth2Provider           *oauth2Provider           `mapstructure:"private-oauth2-provider"`
	OAuth2Providers                 []*oauth2Provider         `mapstructure:"oauth2-providers" wrapstruct:"-"`
	oauth2ClientCredentials         *clientcredentials.Config `wrapstruct:"OAuth2ClientCredentials"`
	modTime                         time.Time
}

func (c *config) init() {
	c.PrivateProviderID = "hydra-sample"
	c.PrivateProviderIDTrustedContext = "hydra-sample-trusted"
	c.PrivateOAuth2Provider.ID = c.PrivateProviderID

	c.OAuth2State.init()

	c.oauth2ClientCredentials = &clientcredentials.Config{
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
	p.oauth2Config = c.newOAuth2Config(p, false)
	p.privateOAuth2Config = c.newOAuth2Config(p, true)

	for _, p := range c.OAuth2Providers {
		p.oauth2Config = c.newOAuth2Config(p, false)
	}

	c.modTime = time.Now()
}

func (c *config) newOAuth2Config(
	p *oauth2Provider,
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
	return &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Scopes:       p.Scopes,
		Endpoint:     endpoint,
		RedirectURL:  c.OAuth2RedirectURL,
	}
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

type oauth2State struct {
	TokenIssuer     string        `mapstructure:"token-issuer"`
	TokenSignKeyHex string        `mapstructure:"token-sign-key" wrapstruct:"-"`
	Expiration      time.Duration `mapstructure:"expiration"`
	tokenSignKey    []byte
}

func (s *oauth2State) init() {
	var err error
	s.tokenSignKey, err = hex.DecodeString(s.TokenSignKeyHex)
	if err != nil {
		panic(err)
	}
}

type oauth2Provider struct {
	ID                  string               `mapstructure:"id"`
	Name                string               `mapstructure:"name"`
	ClientID            string               `mapstructure:"client-id"`
	ClientSecret        string               `mapstructure:"client-secret"`
	PublicKey           string               `mapstructure:"public-key"`
	Scopes              []string             `mapstructure:"scopes"`
	IconURL             string               `mapstructure:"icon"`
	TokenURL            string               `mapstructure:"token-url"`
	AuthURL             string               `mapstructure:"auth-url"`
	oauth2Config        authkit.OAuth2Config `wrapstruct:"-"`
	privateOAuth2Config authkit.OAuth2Config `wrapstruct:"-"`
}

//go:generate wrapstruct -src oauth2Provider -dst oauth2ProviderWrapper -o provider_generated.go
//go:generate wrapstruct -src oauth2State -dst oauth2StateWrapper -o state_generated.go
//go:generate wrapstruct -src config -dst configWrapper -o config_generated.go

type Config struct {
	*configWrapper
}

func (c Config) OAuth2Providers() chan authkit.OAuth2Provider {
	ch := make(chan authkit.OAuth2Provider)
	go func() {
		for _, p := range c.w.OAuth2Providers {
			p := p
			ch <- &oauth2ProviderWrapperEx{&oauth2ProviderWrapper{p}}
		}
		close(ch)
	}()
	return ch
}

func (c Config) OAuth2ProviderByID(id string) authkit.OAuth2Provider {
	for _, p := range c.w.OAuth2Providers {
		if p.ID == id {
			p := p
			return &oauth2ProviderWrapperEx{&oauth2ProviderWrapper{p}}
		}
	}
	return nil
}

func (c Config) OAuth2State() authkit.OAuth2State {
	return &oauth2StateWrapper{&c.w.OAuth2State}
}

func (c Config) PrivateOAuth2Provider() authkit.OAuth2Provider {
	p := c.w.PrivateOAuth2Provider
	return &oauth2ProviderWrapperEx{&oauth2ProviderWrapper{p}}
}

type oauth2ProviderWrapperEx struct {
	*oauth2ProviderWrapper
}

func (p oauth2ProviderWrapperEx) OAuth2Config() authkit.OAuth2Config {
	return p.w.oauth2Config
}

func (p oauth2ProviderWrapperEx) PrivateOAuth2Config() authkit.OAuth2Config {
	return p.w.privateOAuth2Config
}
