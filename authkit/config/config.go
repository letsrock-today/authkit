package config

import (
	"time"

	"golang.org/x/oauth2"
)

type (
	Config interface {
		OAuth2State() OAuth2State
		OAuth2Configs() map[string]oauth2.Config
		OAuth2Providers() []OAuth2Provider
		PrivateOAuth2Provider() OAuth2Provider
		PrivateOAuth2Config() oauth2.Config
		ModTime() time.Time
	}

	OAuth2State interface {
		TokenIssuer() string
		TokenSignKey() []byte
		Expiration() time.Duration
	}

	OAuth2Provider interface {
		ID() string
		Name() string
		ClientID() string
		ClientSecret() string
		PublicKey() string
		Scopes() []string
		IconURL() string
		TokenURL() string
		AuthURL() string
	}
)
