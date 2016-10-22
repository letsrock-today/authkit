package config

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/spf13/viper"

	"github.com/letsrock-today/hydra-sample/authkit"
)

const (
	defPath = "./env"
	defName = "dev"
)

// Init initializes app's config.
// App should invoke this function from the main after it parsed flags.
// prefPath, prefName allows to overwrite default values for config dir and base file name.
func Init(prefPath, prefName string) {
	if prefPath != "" {
		viper.AddConfigPath(prefPath)
	}
	viper.AddConfigPath(defPath)
	viper.AddConfigPath(".")
	if prefName == "" {
		prefName = defName
	}
	viper.SetConfigName(prefName)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func Get() Config {
	var c Config
	err := viper.Unmarshal(&c)
	if err != nil {
		panic(err)
	}
	c.PrivateProviderID = "hydra-sample"
	c.PrivateProviderIDTrustedContext = "hydra-sample-trusted"
	return c
}

type Config struct {
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

	authCookieName string `mapstructure:"auth-cookie-name"`

	EmailConfig struct {
		Sender     string `mapstructure:"sender"`
		SenderPass string `mapstructure:"sender-pass"`
		MailServer string `mapstructure:"server"`
		MailPort   string `mapstructure:"port"`
	} `mapstructure:"email-config"`

	MongoDB struct {
		URL  string `mapstructure:"url"`
		Name string `mapstructure:"name"`
	} `mapstructure:"mongodb"`

	oauth2State           oauth2State      `mapstructure:"oauth2-state"`
	privateOAuth2Provider oauth2Provider   `mapstructure:"private-oauth2-provider"`
	oauth2Providers       []oauth2Provider `mapstructure:"oauth2-providers"`
	modTime               time.Time        `mapstructure:"-"`
}

func (c Config) AuthCookieName() string {
	return c.authCookieName
}

func (c Config) OAuth2State() authkit.OAuth2State {
	return c.oauth2State
}

func (c Config) PrivateOAuth2Provider() authkit.OAuth2Provider {
	c.privateOAuth2Provider.c = &c
	return c.privateOAuth2Provider
}

func (c Config) OAuth2ClientCredentials() *clientcredentials.Config {
	return &clientcredentials.Config{
		ClientID:     c.privateOAuth2Provider.clientID,
		ClientSecret: c.privateOAuth2Provider.clientSecret,
		Scopes:       c.privateOAuth2Provider.scopes,
		TokenURL: strings.Replace(
			c.privateOAuth2Provider.tokenURL,
			"{base-url}",
			c.HydraAddr,
			-1),
	}
}

func (c Config) OAuth2Providers() chan authkit.OAuth2Provider {
	ch := make(chan authkit.OAuth2Provider)
	go func() {
		for _, p := range c.oauth2Providers {
			p.c = &c
			ch <- p
		}
		close(ch)
	}()
	return ch
}

func (c Config) OAuth2ProviderByID(id string) authkit.OAuth2Provider {
	for _, p := range c.oauth2Providers {
		if p.ID() == id {
			p.c = &c
			return p
		}
	}
	return nil
}

func (c Config) ModTime() time.Time {
	return c.modTime
}

type oauth2Provider struct {
	id           string   `mapstructure:"id"`
	name         string   `mapstructure:"name"`
	clientID     string   `mapstructure:"client-id"`
	clientSecret string   `mapstructure:"client-secret"`
	publicKey    string   `mapstructure:"public-key"`
	scopes       []string `mapstructure:"scopes"`
	iconURL      string   `mapstructure:"icon"`
	tokenURL     string   `mapstructure:"token-url"`
	authURL      string   `mapstructure:"auth-url"`
	c            *Config  `mapstructure:"-"`
}

func (p oauth2Provider) ID() string {
	return p.id
}

func (p oauth2Provider) Name() string {
	return p.name
}

func (p oauth2Provider) IconURL() string {
	return p.iconURL
}

func (p oauth2Provider) OAuth2Config() authkit.OAuth2Config {
	return newOAuth2Config(p, false)
}

func (p oauth2Provider) PrivateOAuth2Config() authkit.OAuth2Config {
	return newOAuth2Config(p, true)
}

func newOAuth2Config(p oauth2Provider, private bool) authkit.OAuth2Config {
	var baseURL string
	if private {
		baseURL = p.c.HydraAddr
	} else {
		baseURL = p.c.ExternalBaseURL
	}
	tokenURL := strings.Replace(p.tokenURL, "{base-url}", baseURL, -1)
	authURL := strings.Replace(p.authURL, "{base-url}", baseURL, -1)
	return &oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		Scopes:       p.scopes,
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenURL,
			AuthURL:  authURL,
		},
		RedirectURL: viper.GetString("oauth2-redirect-url"),
	}
}

type oauth2State struct {
	tokenIssuer     string        `mapstructure:"token-issuer"`
	tokenSignKeyHex string        `mapstructure:"token-sign-key"`
	expiration      time.Duration `mapstructure:"expiration"`
}

func (s oauth2State) TokenIssuer() string {
	return s.tokenIssuer
}

func (s oauth2State) TokenSignKey() []byte {
	tokenSignKey, err := hex.DecodeString(s.tokenSignKeyHex)
	if err != nil {
		panic(err)
	}
	return tokenSignKey
}

func (s oauth2State) Expiration() time.Duration {
	return s.expiration
}
