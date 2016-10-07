package handler

import (
	"time"

	"github.com/letsrock-today/hydra-sample/authkit/config"
	"golang.org/x/oauth2"
)

type testConfig struct {
	oauth2State     testOAuth2State
	oauth2Configs   map[string]oauth2.Config
	oauth2Providers []testOAuth2Provider
	modTime         time.Time
}

func (c testConfig) OAuth2State() config.OAuth2State {
	return c.oauth2State
}

func (c testConfig) OAuth2Configs() map[string]oauth2.Config {
	return c.oauth2Configs
}

func (c testConfig) OAuth2Providers() []config.OAuth2Provider {
	r := []config.OAuth2Provider{}
	for _, p := range c.oauth2Providers {
		r = append(r, p)
	}
	return r
}

func (c testConfig) ModTime() time.Time {
	return c.modTime
}

type testOAuth2State struct {
	tokenIssuer  string
	tokenSignKey []byte
	expiration   time.Duration
}

func (s testOAuth2State) TokenIssuer() string {
	return s.tokenIssuer
}

func (s testOAuth2State) TokenSignKey() []byte {
	return s.tokenSignKey
}

func (s testOAuth2State) Expiration() time.Duration {
	return s.expiration
}

type testOAuth2Provider struct {
	id           string
	name         string
	clientId     string
	clientSecret string
	publicKey    string
	scopes       []string
	iconURL      string
	tokenURL     string
	authURL      string
}

func (p testOAuth2Provider) ID() string {
	return p.id
}

func (p testOAuth2Provider) Name() string {
	return p.name
}

func (p testOAuth2Provider) ClientId() string {
	return p.clientId
}

func (p testOAuth2Provider) ClientSecret() string {
	return p.clientSecret
}

func (p testOAuth2Provider) PublicKey() string {
	return p.publicKey
}

func (p testOAuth2Provider) Scopes() []string {
	return p.scopes
}

func (p testOAuth2Provider) IconURL() string {
	return p.iconURL
}

func (p testOAuth2Provider) TokenURL() string {
	return p.tokenURL
}

func (p testOAuth2Provider) AuthURL() string {
	return p.authURL
}
