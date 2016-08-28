package oauth2

import (
	"errors"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type OAuth2Config interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	PasswordCredentialsToken(ctx context.Context, username, password string) (*oauth2.Token, error)
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
	TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource
	GetAuthCodeOptions(v url.Values) []oauth2.AuthCodeOption
}

type Config struct {
	*oauth2.Config
}

func (c *Config) GetAuthCodeOptions(v url.Values) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{oauth2.AccessTypeOnline}
}

func (c *Config) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	return &tokenSource{c.Config.TokenSource(ctx, t), t, ctx}
}

type tokenSource struct {
	oauth2.TokenSource
	token *oauth2.Token
	ctx   context.Context
}

func (s *tokenSource) Token() (*oauth2.Token, error) {
	newToken, err := s.TokenSource.Token()
	if err != nil {
		return nil, err
	}
	if newToken != s.token {
		cb := s.ctx.Value("SaveTokenCallback")
		if cb != nil {
			if f, ok := cb.(func(*oauth2.Token)); ok {
				f(newToken)
			} else {
				return nil, errors.New("Illegal SaveTokenCallback in context")
			}
		}
		s.token = newToken
	}
	return newToken, nil
}
