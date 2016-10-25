package config

import (
	"encoding/hex"
	"strings"
	"time"

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

	"github.com/pkg/errors"

	"github.com/letsrock-today/hydra-sample/authkit"
)

const (
	defPath = "./env"
	defName = "dev"
)

var (
	c         Config
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

type config struct {
	PrivateProviderID               string `mapstructure:"-"`
	PrivateProviderIDTrustedContext string `mapstructure:"-"`

	ListenAddr            string `mapstructure:"listen-addr"`
	TLSCertFile           string `mapstructure:"tls-cert-file"`
	TLSKeyFile            string `mapstructure:"tls-key-file"`
	TLSInsecureSkipVerify bool   `mapstructure:"tls-insecure-skip-verify"`

	HydraAddr         string `mapstructure:"hydra-addr"`
	ExternalBaseURL   string `mapstructure:"external-base-url"`
	OAuth2RedirectURL string `mapstructure:"oauth2-redirect-url"`

	ChallengeLifespan        time.Duration `mapstructure:"challenge-lifespan"`
	ConfirmationLinkLifespan time.Duration `mapstructure:"confirmation-link-lifespan"`

	AuthCookieName string `mapstructure:"auth-cookie-name"`

	EmailConfig EmailConfig `mapstructure:"email-config"`

	MongoDB DBConfig `mapstructure:"mongodb"`

	OAuth2State           oauth2State      `mapstructure:"oauth2-state"`
	PrivateOAuth2Provider *oauth2Provider  `mapstructure:"private-oauth2-provider"`
	OAuth2Providers       []oauth2Provider `mapstructure:"oauth2-providers"`
	modTime               time.Time
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
	TokenSignKeyHex string        `mapstructure:"token-sign-key"`
	Expiration      time.Duration `mapstructure:"expiration"`
}

type oauth2Provider struct {
	ID           string   `mapstructure:"id"`
	Name         string   `mapstructure:"name"`
	ClientID     string   `mapstructure:"client-id"`
	ClientSecret string   `mapstructure:"client-secret"`
	PublicKey    string   `mapstructure:"public-key"`
	Scopes       []string `mapstructure:"scopes"`
	IconURL      string   `mapstructure:"icon"`
	TokenURL     string   `mapstructure:"token-url"`
	AuthURL      string   `mapstructure:"auth-url"`
}

// We have had to wrap structure into interface here, because mixture of interface
// and struct felt awful. Package authkit uses interface to abstract away from
// application's config, so we need to implement authkit.Config somewhere.
type Config struct {
	c *config
}

func (c Config) PrivateProviderID() string {
	return c.c.PrivateProviderID
}

func (c Config) PrivateProviderIDTrustedContext() string {
	return c.c.PrivateProviderIDTrustedContext
}

func (c Config) ListenAddr() string {
	return c.c.ListenAddr
}

func (c Config) TLSCertFile() string {
	return c.c.TLSCertFile
}

func (c Config) TLSKeyFile() string {
	return c.c.TLSKeyFile
}

func (c Config) TLSInsecureSkipVerify() bool {
	return c.c.TLSInsecureSkipVerify
}

func (c Config) HydraAddr() string {
	return c.c.HydraAddr
}

func (c Config) ExternalBaseURL() string {
	return c.c.ExternalBaseURL
}

func (c Config) OAuth2RedirectURL() string {
	return c.c.OAuth2RedirectURL
}

func (c Config) ChallengeLifespan() time.Duration {
	return c.c.ChallengeLifespan
}

func (c Config) ConfirmationLinkLifespan() time.Duration {
	return c.c.ConfirmationLinkLifespan
}

func (c Config) AuthCookieName() string {
	return c.c.AuthCookieName
}

func (c Config) EmailConfig() EmailConfig {
	return c.c.EmailConfig
}

func (c Config) MongoDB() DBConfig {
	return c.c.MongoDB
}

func (c Config) OAuth2State() authkit.OAuth2State {
	tokenSignKey, err := hex.DecodeString(c.c.OAuth2State.TokenSignKeyHex)
	if err != nil {
		panic(err)
	}
	return oauth2StateImpl{c.c.OAuth2State, tokenSignKey}
}

func (c Config) PrivateOAuth2Provider() authkit.OAuth2Provider {
	p := c.c.PrivateOAuth2Provider
	return oauth2ProviderImpl{
		p,
		newOAuth2Config(c.c, p, false),
		newOAuth2Config(c.c, p, true),
	}
}

func (c Config) OAuth2ClientCredentials() *clientcredentials.Config {
	return &clientcredentials.Config{
		ClientID:     c.c.PrivateOAuth2Provider.ClientID,
		ClientSecret: c.c.PrivateOAuth2Provider.ClientSecret,
		Scopes:       c.c.PrivateOAuth2Provider.Scopes,
		TokenURL: strings.Replace(
			c.c.PrivateOAuth2Provider.TokenURL,
			"{base-url}",
			c.c.HydraAddr,
			-1),
	}
}

func (c Config) OAuth2Providers() chan authkit.OAuth2Provider {
	ch := make(chan authkit.OAuth2Provider)
	go func() {
		for _, p := range c.c.OAuth2Providers {
			p := p
			ch <- oauth2ProviderImpl{
				&p,
				newOAuth2Config(c.c, &p, false),
				nil,
			}
		}
		close(ch)
	}()
	return ch
}

func (c Config) OAuth2ProviderByID(id string) authkit.OAuth2Provider {
	for _, p := range c.c.OAuth2Providers {
		if p.ID == id {
			p := p
			return oauth2ProviderImpl{
				&p,
				newOAuth2Config(c.c, &p, false),
				nil,
			}
		}
	}
	return nil
}

func (c Config) ModTime() time.Time {
	return c.c.modTime
}

type oauth2ProviderImpl struct {
	p                   *oauth2Provider
	oauth2Config        authkit.OAuth2Config
	privateOAuth2Config authkit.OAuth2Config
}

func (p oauth2ProviderImpl) ID() string {
	return p.p.ID
}

func (p oauth2ProviderImpl) Name() string {
	return p.p.Name
}

func (p oauth2ProviderImpl) IconURL() string {
	return p.p.IconURL
}

func (p oauth2ProviderImpl) OAuth2Config() authkit.OAuth2Config {
	return p.oauth2Config
}

func (p oauth2ProviderImpl) PrivateOAuth2Config() authkit.OAuth2Config {
	return p.privateOAuth2Config
}

func newOAuth2Config(c *config, p *oauth2Provider, privateConfig bool) authkit.OAuth2Config {
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
			panic(errors.Errorf("Illegal OAuth2 configuration, for: %s, %t", p.ID, privateConfig))
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

type oauth2StateImpl struct {
	s            oauth2State
	tokenSignKey []byte
}

func (s oauth2StateImpl) TokenIssuer() string {
	return s.s.TokenIssuer
}

func (s oauth2StateImpl) TokenSignKey() []byte {
	return s.tokenSignKey
}

func (s oauth2StateImpl) Expiration() time.Duration {
	return s.s.Expiration
}
