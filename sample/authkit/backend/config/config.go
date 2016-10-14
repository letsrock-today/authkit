package config

import (
	"log"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/letsrock-today/hydra-sample/authkit/config"
)

const (
	PrivPID = "hydra-sample"
)

type Config struct {
	ListenAddr               string                   `yaml:"listen-addr"`
	TLSCertFile              string                   `yaml:"tls-cert-file"`
	TLSKeyFile               string                   `yaml:"tls-key-file"`
	ExternalBaseURL          string                   `yaml:"external-base-url"`
	OAuth2RedirectURL        string                   `yaml:"oauth2-redirect-url"`
	HydraOAuth2Provider      OAuth2Provider           `yaml:"hydra-clientcredentials"`
	OAuth2State              OAuth2State              `yaml:"oauth2-state"`
	OAuth2Providers          []OAuth2Provider         `yaml:"oauth2-providers"`
	HydraAddr                string                   `yaml:"hydra-addr"`
	ChallengeLifespan        time.Duration            `yaml:"challenge-lifespan"`
	ConfirmationLinkLifespan time.Duration            `yaml:"confirmation-link-lifespan"`
	EmailConfig              EmailConfig              `yaml:"email-config"`
	HydraClientCredentials   clientcredentials.Config `yaml:"-"`
	HydraOAuth2Config        oauth2.Config            `yaml:"-"`
	HydraOAuth2ConfigInt     oauth2.Config            `yaml:"-"`
	OAuth2Configs            map[string]oauth2.Config `yaml:"-"`
	modTime                  time.Time                `yaml:"-"`
	TLSInsecureSkipVerify    bool                     `yaml:"tls-insecure-skip-verify"`
}

type OAuth2State struct {
	TokenIssuer     string        `yaml:"token-issuer"`
	TokenSignKeyHex string        `yaml:"token-sign-key"`
	TokenSignKey    []byte        `yaml:"-"`
	Expiration      time.Duration `yaml:"expiration"`
}

type OAuth2Provider struct {
	Id           string   `json:"id" yaml:"id"`
	Name         string   `json:"name" yaml:"name"`
	ClientId     string   `json:"-" yaml:"client-id"`
	ClientSecret string   `json:"-" yaml:"client-secret"`
	PublicKey    string   `json:"-" yaml:"public-key"`
	Scopes       []string `json:"-" yaml:"scopes"`
	IconURL      string   `json:"iconUrl" yaml:"icon"`
	TokenURL     string   `json:"-" yaml:"token-url"`
	AuthURL      string   `json:"-" yaml:"auth-url"`
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

func GetCfg() config.Config {
	return _config{Get()}
}

type _config struct {
	c Config
}

func (c _config) OAuth2State() config.OAuth2State {
	return _oauth2State{c.c.OAuth2State}
}

func (c _config) OAuth2Configs() map[string]oauth2.Config {
	return c.c.OAuth2Configs
}

func (c _config) OAuth2Providers() []config.OAuth2Provider {
	pp := []config.OAuth2Provider{}
	for _, p := range c.c.OAuth2Providers {
		pp = append(pp, _oauth2Provider{p})
	}
	return pp
}

func (c _config) PrivateOAuth2Config() oauth2.Config {
	return c.c.HydraOAuth2Config
}

func (c _config) PrivateOAuth2Provider() config.OAuth2Provider {
	return _oauth2Provider{c.c.HydraOAuth2Provider}
}

func (c _config) ModTime() time.Time {
	return c.c.modTime
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
