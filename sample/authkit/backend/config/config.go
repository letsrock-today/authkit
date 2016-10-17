package config

import (
	"log"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/letsrock-today/hydra-sample/authkit"
)

const (
	PrivPID = "hydra-sample"
)

type Config struct {
	ListenAddr               string                   `yaml:"listen-addr"`
	TLSCertFile              string                   `yaml:"tls-cert-file"`
	TLSKeyFile               string                   `yaml:"tls-key-file"`
	HydraAddr                string                   `yaml:"hydra-addr"`
	ExternalBaseURL          string                   `yaml:"external-base-url"`
	OAuth2RedirectURL        string                   `yaml:"oauth2-redirect-url"`
	OAuth2State              OAuth2State              `yaml:"oauth2-state"`
	HydraClientCredentials   clientcredentials.Config `yaml:"-"`
	HydraOAuth2Provider      OAuth2Provider           `yaml:"hydra-clientcredentials"`
	OAuth2Providers          []OAuth2Provider         `yaml:"oauth2-providers"`
	ChallengeLifespan        time.Duration            `yaml:"challenge-lifespan"`
	ConfirmationLinkLifespan time.Duration            `yaml:"confirmation-link-lifespan"`
	EmailConfig              EmailConfig              `yaml:"email-config"`
	modTime                  time.Time                `yaml:"-"`
	TLSInsecureSkipVerify    bool                     `yaml:"tls-insecure-skip-verify"`
	AuthCookieName           string                   `yaml:"auth-cookie-name"`
}

type OAuth2State struct {
	TokenIssuer     string        `yaml:"token-issuer"`
	TokenSignKeyHex string        `yaml:"token-sign-key"`
	TokenSignKey    []byte        `yaml:"-"`
	Expiration      time.Duration `yaml:"expiration"`
}

type OAuth2Provider struct {
	Id                  string         `json:"id" yaml:"id"`
	Name                string         `json:"name" yaml:"name"`
	ClientId            string         `json:"-" yaml:"client-id"`
	ClientSecret        string         `json:"-" yaml:"client-secret"`
	PublicKey           string         `json:"-" yaml:"public-key"`
	Scopes              []string       `json:"-" yaml:"scopes"`
	IconURL             string         `json:"iconUrl" yaml:"icon"`
	TokenURL            string         `json:"-" yaml:"token-url"`
	AuthURL             string         `json:"-" yaml:"auth-url"`
	OAuth2Config        *oauth2.Config `json:"-" yaml:"-"`
	PrivateOAuth2Config *oauth2.Config `json:"-" yaml:"-"`
}

type EmailConfig struct {
	Sender     string `yaml:"sender"`
	SenderPass string `yaml:"sender-pass"`
	MailServer string `yaml:"server"`
	MailPort   string `yaml:"port"`
}

func Get() Config {
	once.Do(func() {
		log.Printf("Program started with config file '%s'", cfgPath)
		cfg.Store(loadConfig())
		initConfigFileWatcher()
	})
	return cfg.Load().(Config)
}

func ModTime() time.Time {
	return Get().modTime
}

//////////////////////

//TODO: this is temporary adapter, refactoring needed

func GetCfg() authkit.Config {
	return _config{Get()}
}

type _config struct {
	c Config
}

func (c _config) OAuth2State() authkit.OAuth2State {
	return _oauth2State{c.c.OAuth2State}
}

func (c _config) OAuth2Providers() chan authkit.OAuth2Provider {
	ch := make(chan authkit.OAuth2Provider)
	go func() {
		for _, p := range c.c.OAuth2Providers {
			ch <- _oauth2Provider{p}
		}
		close(ch)
	}()
	return ch
}

func (c _config) PrivateOAuth2Provider() authkit.OAuth2Provider {
	return _oauth2Provider{c.c.HydraOAuth2Provider}
}

func (c _config) OAuth2ProviderByID(id string) authkit.OAuth2Provider {
	for _, p := range c.c.OAuth2Providers {
		if p.Id == id {
			return _oauth2Provider{p}
		}
	}
	return nil
}

func (c _config) ModTime() time.Time {
	return c.c.modTime
}

func (c _config) TLSInsecureSkipVerify() bool {
	return c.c.TLSInsecureSkipVerify
}

func (c _config) AuthCookieName() string {
	return c.c.AuthCookieName
}

type _oauth2State struct {
	s OAuth2State
}

func (s _oauth2State) TokenIssuer() string {
	return s.s.TokenIssuer
}

func (s _oauth2State) TokenSignKey() []byte {
	return s.s.TokenSignKey
}

func (s _oauth2State) Expiration() time.Duration {
	return s.s.Expiration
}

type _oauth2Provider struct {
	p OAuth2Provider
}

func (p _oauth2Provider) ID() string {
	return PrivPID
}

func (p _oauth2Provider) Name() string {
	return p.p.Name
}

func (p _oauth2Provider) ClientID() string {
	return p.p.ClientId
}

func (p _oauth2Provider) ClientSecret() string {
	return p.p.ClientSecret
}

func (p _oauth2Provider) PublicKey() string {
	return p.p.PublicKey
}

func (p _oauth2Provider) Scopes() []string {
	return p.p.Scopes
}

func (p _oauth2Provider) IconURL() string {
	return p.p.IconURL
}

func (p _oauth2Provider) TokenURL() string {
	return p.p.TokenURL
}

func (p _oauth2Provider) AuthURL() string {
	return p.p.AuthURL
}

func (p _oauth2Provider) OAuth2Config() authkit.OAuth2Config {
	return p.p.OAuth2Config
}

func (p _oauth2Provider) PrivateOAuth2Config() authkit.OAuth2Config {
	return p.p.PrivateOAuth2Config
}
