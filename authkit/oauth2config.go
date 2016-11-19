package authkit

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type (
	// OAuth2State holds configuration parameters for OAuth2 code flow state token.
	OAuth2State struct {
		TokenIssuer  string
		TokenSignKey []byte
		Expiration   time.Duration
	}

	// OAuth2Provider holds OAuth2 provider's configuraton.
	OAuth2Provider struct {
		ID           string
		Name         string
		IconURL      string
		OAuth2Config OAuth2Config

		// PrivateOAuth2Config used to access private provider via private network.
		// So, URLs may be accessible only within DMZ, hence different config.
		PrivateOAuth2Config OAuth2Config
	}

	// OAuth2Config is an interface extracted from the "golang.org/x/oauth2".Config.
	// This interface extracted for testability.
	OAuth2Config interface {
		TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource
		AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
		PasswordCredentialsToken(ctx context.Context, username, password string) (*oauth2.Token, error)
		Exchange(ctx context.Context, code string) (*oauth2.Token, error)
		Client(ctx context.Context, t *oauth2.Token) *http.Client
	}
)

//go:generate mockery -name OAuth2Config
