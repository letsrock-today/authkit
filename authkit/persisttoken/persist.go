package persisttoken

import (
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/letsrock-today/hydra-sample/authkit"
)

// WrapOAuth2Config wraps authkit.OAuth2Config or oauth2.Config with
// logic for store/retrieve token from provided authkit.TokenStore.
func WrapOAuth2Config(
	c authkit.OAuth2Config,
	login, providerID string,
	ts authkit.TokenStore) authkit.OAuth2Config {
	return &config{
		cfg:        c,
		login:      login,
		providerID: providerID,
		ts:         ts,
	}
}

type config struct {
	cfg        authkit.OAuth2Config
	login      string
	providerID string
	ts         authkit.TokenStore
}

func (c *config) Client(
	ctx context.Context,
	t *oauth2.Token) *http.Client {
	return oauth2.NewClient(ctx, c.TokenSource(ctx, t))
}

func (c *config) TokenSource(
	ctx context.Context,
	t *oauth2.Token) oauth2.TokenSource {
	return oauth2.ReuseTokenSource(
		t,
		persistTokenSource{
			login:      c.login,
			providerID: c.providerID,
			ts:         c.ts,
			cfg:        c.cfg,
			ctx:        ctx,
		})
}

func (c *config) AuthCodeURL(
	state string,
	opts ...oauth2.AuthCodeOption) string {
	return c.cfg.AuthCodeURL(state, opts...)
}

func (c *config) PasswordCredentialsToken(
	ctx context.Context,
	username, password string) (*oauth2.Token, error) {
	return c.cfg.PasswordCredentialsToken(ctx, username, password)
}

func (c *config) Exchange(
	ctx context.Context,
	code string) (*oauth2.Token, error) {
	return c.cfg.Exchange(ctx, code)
}

type persistTokenSource struct {
	login      string
	providerID string
	ts         authkit.TokenStore
	cfg        authkit.OAuth2Config
	ctx        context.Context
}

func (p persistTokenSource) Token() (*oauth2.Token, error) {
	t, err := p.ts.OAuth2Token(p.login, p.providerID)
	if err != nil {
		return nil, err
	}
	new, err := p.cfg.TokenSource(p.ctx, t).Token()
	if err != nil {
		return nil, err
	}
	if new != t {
		return new, p.ts.UpdateOAuth2Token(p.login, p.providerID, new)
	}
	return new, nil
}
