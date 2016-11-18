package persisttoken

import (
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/letsrock-today/authkit/authkit"
)

// WrapOAuth2Config wraps authkit.OAuth2Config or oauth2.Config with
// logic for store/retrieve token from provided authkit.TokenStore.
func WrapOAuth2Config(
	c authkit.OAuth2Config,
	login, providerID string,
	ts authkit.TokenStore,
	checkRefreshable func(*oauth2.Token) error) authkit.OAuth2Config {
	return &config{
		cfg:        c,
		providerID: providerID,
		ts:         ts,
		prepare: func() (*oauth2.Token, string, error) {
			t, err := ts.OAuth2Token(login, providerID)
			return t, login, err
		},
		checkRefreshable: checkRefreshable,
	}
}

// WrapOAuth2ConfigUseAccessToken wraps authkit.OAuth2Config or oauth2.Config
// with logic for store/retrieve token from provided authkit.TokenStore.
// It uses passed access token to find associated login and OAuth2 token.
func WrapOAuth2ConfigUseAccessToken(
	c authkit.OAuth2Config,
	accessToken, providerID string,
	ts authkit.TokenStore,
	checkRefreshable func(*oauth2.Token) error) authkit.OAuth2Config {
	return &config{
		cfg:        c,
		providerID: providerID,
		ts:         ts,
		prepare: func() (*oauth2.Token, string, error) {
			return ts.OAuth2TokenAndLoginByAccessToken(
				accessToken,
				providerID)
		},
		checkRefreshable: checkRefreshable,
	}
}

type config struct {
	cfg              authkit.OAuth2Config
	providerID       string
	ts               authkit.TokenStore
	prepare          func() (*oauth2.Token, string, error)
	checkRefreshable func(*oauth2.Token) error
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
			providerID:       c.providerID,
			ts:               c.ts,
			cfg:              c.cfg,
			ctx:              ctx,
			prepare:          c.prepare,
			checkRefreshable: c.checkRefreshable,
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
	providerID       string
	ts               authkit.TokenStore
	cfg              authkit.OAuth2Config
	ctx              context.Context
	prepare          func() (*oauth2.Token, string, error)
	checkRefreshable func(*oauth2.Token) error
}

func (p persistTokenSource) Token() (*oauth2.Token, error) {
	t, login, err := p.prepare()
	if err != nil {
		return nil, err
	}
	new, err := p.cfg.TokenSource(p.ctx, t).Token()
	if err != nil {
		return nil, err
	}
	if new != t {
		if err := p.ts.UpdateOAuth2Token(login, p.providerID, new); err != nil {
			return nil, err
		}
		// Still refresh token for server's internal use, but return error to
		// the caller, if it provided validation function. This is useful to
		// emulate session timeout for web UI.
		if p.checkRefreshable != nil {
			if err := p.checkRefreshable(t); err != nil {
				return nil, err
			}
		}
	}
	return new, nil
}
