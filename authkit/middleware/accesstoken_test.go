package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidation(t *testing.T) {
	assert := assert.New(t)

	// No config at all
	// No obligatory settings
	assert.Panics(func() {
		AccessTokenWithConfig(AccessTokenConfig{})
	})

	// Ensure that defaults are used if non-obligatory settings are not set
	var effcfg *AccessTokenConfig = nil
	reportEffectiveConfig = func(c AccessTokenConfig) {
		effcfg = &c
	}
	AccessTokenWithConfig(AccessTokenConfig{
		UserStore: &userStore{
			accessToken: "xxx",
			user:        user{Email: "email", Name: "name"},
		},
		TokenValidator: tokenValidator{},
	})
	assert.NotNil(effcfg)
	assert.Equal(DefaultContextKey, effcfg.ContextKey)
	assert.Equal(NewDefaultPermissionMapper(), effcfg.PermissionMapper)

	// Ensure that defaults are used in AccessToken except fields set explicitly
	effcfg = nil
	AccessToken(
		&userStore{
			accessToken: "xxx",
			user:        user{Email: "email", Name: "name"},
		},
		tokenValidator{})
	assert.NotNil(effcfg)
	assert.Equal(DefaultContextKey, effcfg.ContextKey)
	assert.Equal(NewDefaultPermissionMapper(), effcfg.PermissionMapper)
	assert.NotNil(effcfg.UserStore)
	assert.NotNil(effcfg.TokenValidator)
}

func TestAccessTokenWithConfig(t *testing.T) {
	accessTokenConfig := AccessTokenConfig{
		UserStore: &userStore{
			accessToken: "xxx",
			user:        user{Email: "email", Name: "name"},
		},
		PermissionMapper: permMapper{},
		TokenValidator: tokenValidator{
			allowed: map[string]bool{
				"GET:/permitted:xxx":     true,
				"GET:/permitted:yyy":     true,
				"POST:/permitted:xxx":    true,
				"POST:/permitted:yyy":    true,
				"GET:/permitted-get:xxx": true,
				"GET:/permitted-get:yyy": true,
			},
		},
		OAuth2Config:  tokenSourceProvider{},
		OAuth2Context: context.Background(),
	}

	invalidHeaderFormatMsg := InvalidAuthHeaderError.Message
	notPermittedMsg := AccessDeniedError.Message

	cases := []struct {
		name                 string
		w                    *httptest.ResponseRecorder
		r                    *http.Request
		accessTokenConfig    AccessTokenConfig
		accessTokenHeader    string
		expectedResponseCode int
		expectedResponseBody string
		unprotected          bool
	}{
		{
			name:                 "Unprotected resource, header empty",
			w:                    httptest.NewRecorder(),
			r:                    newGetUnprotected(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "",
			expectedResponseCode: http.StatusOK,
			expectedResponseBody: nextHandlerMsg,
			unprotected:          true,
		},
		{
			name:                 "Protected resource, header empty",
			w:                    httptest.NewRecorder(),
			r:                    newGetRestricted(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "",
			expectedResponseCode: http.StatusForbidden,
			expectedResponseBody: invalidHeaderFormatMsg,
		},
		{
			name:                 "Protected resource, invalid header format (no 'bearer')",
			w:                    httptest.NewRecorder(),
			r:                    newGetRestricted(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "zzz",
			expectedResponseCode: http.StatusForbidden,
			expectedResponseBody: invalidHeaderFormatMsg,
		},
		{
			name:                 "Protected resource, invalid header format (no token after 'bearer')",
			w:                    httptest.NewRecorder(),
			r:                    newGetRestricted(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "bearer   ",
			expectedResponseCode: http.StatusForbidden,
			expectedResponseBody: invalidHeaderFormatMsg,
		},
		{
			name:                 "Restricted resource, valid token",
			w:                    httptest.NewRecorder(),
			r:                    newGetRestricted(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "bearer xxx",
			expectedResponseCode: http.StatusForbidden,
			expectedResponseBody: notPermittedMsg,
		},
		{
			name:                 "Permitted resource, valid token, GET",
			w:                    httptest.NewRecorder(),
			r:                    newGetPermitted(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "bearer xxx",
			expectedResponseCode: http.StatusOK,
			expectedResponseBody: nextHandlerMsg,
		},
		{
			name:                 "Permitted resource, valid token, POST",
			w:                    httptest.NewRecorder(),
			r:                    newPostPermitted(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "bearer xxx",
			expectedResponseCode: http.StatusOK,
			expectedResponseBody: nextHandlerMsg,
		},
		{
			name:                 "Permitted resource, invalid token",
			w:                    httptest.NewRecorder(),
			r:                    newGetPermitted(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "bearer zzz",
			expectedResponseCode: http.StatusForbidden,
			expectedResponseBody: notPermittedMsg,
		},
		{
			name:                 "Permitted resource, unknown token",
			w:                    httptest.NewRecorder(),
			r:                    newGetPermitted(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "bearer yyy", // token with permission, but without user
			expectedResponseCode: http.StatusForbidden,
			expectedResponseBody: notPermittedMsg,
		},
		{
			name:                 "Permitted only get resource, valid token, GET",
			w:                    httptest.NewRecorder(),
			r:                    newGetPermittedOnlyGet(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "bearer xxx",
			expectedResponseCode: http.StatusOK,
			expectedResponseBody: nextHandlerMsg,
		},
		{
			name:                 "Permitted only get resource, valid token, POST",
			w:                    httptest.NewRecorder(),
			r:                    newPostPermittedOnlyGet(t),
			accessTokenConfig:    accessTokenConfig,
			accessTokenHeader:    "bearer xxx",
			expectedResponseCode: http.StatusForbidden,
			expectedResponseBody: notPermittedMsg,
		},
	}

	for _, cs := range cases {
		cs := cs
		// e.Any(...) and brothers should not be used in parallel
		next := nextHandler{
			checkPrincipal: !cs.unprotected,
		}
		e := echo.New()
		e.Any("/unprotected", next.next)
		e.Any(
			"/permitted",
			next.next,
			AccessTokenWithConfig(cs.accessTokenConfig))
		e.Any(
			"/permitted-get",
			next.next,
			AccessTokenWithConfig(cs.accessTokenConfig))
		e.Any(
			"/restricted",
			next.next,
			AccessTokenWithConfig(cs.accessTokenConfig))
		t.Run(cs.name, func(st *testing.T) {
			st.Parallel()
			assert := assert.New(st)
			r := cs.r
			w := cs.w
			cs.r.Header.Set("Authorization", cs.accessTokenHeader)
			e.ServeHTTP(
				standard.NewRequest(r, e.Logger()),
				standard.NewResponse(w, e.Logger()))
			assert.Equal(cs.expectedResponseCode, w.Code)
			assert.Equal(cs.expectedResponseBody, string(w.Body.Bytes()))
			if !cs.unprotected && w.Code == http.StatusOK {
				assert.True(cs.accessTokenConfig.UserStore.(*userStore).tokenRefreshed)
			}
		})
	}
}

func newGetUnprotected(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.GET, "/unprotected", nil)
	assert.NoError(t, err)
	return r
}

func newGetPermitted(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.GET, "/permitted", nil)
	assert.NoError(t, err)
	return r
}

func newPostPermitted(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.POST, "/permitted", nil)
	assert.NoError(t, err)
	return r
}

func newGetPermittedOnlyGet(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.GET, "/permitted-get", nil)
	assert.NoError(t, err)
	return r
}

func newPostPermittedOnlyGet(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.POST, "/permitted-get", nil)
	assert.NoError(t, err)
	return r
}

func newGetRestricted(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.GET, "/restricted", nil)
	assert.NoError(t, err)
	return r
}

const nextHandlerMsg = "Result from next handler"

type nextHandler struct {
	hasRun         bool
	checkPrincipal bool
}

func (n *nextHandler) next(c echo.Context) error {
	n.hasRun = true
	// check if user data is available in context
	if !n.checkPrincipal {
		return c.String(http.StatusOK, nextHandlerMsg)
	}
	u := c.Get(DefaultContextKey)
	if u, ok := u.(user); !ok {
		return errors.New("no user in context")
	} else {
		if u.Name != "name" {
			return errors.New("invalid user in context")
		}
	}
	return c.String(http.StatusOK, nextHandlerMsg)
}

type user struct {
	Email string
	Name  string
}

type userStore struct {
	accessToken    string
	user           user
	tokenRefreshed bool
}

func (s *userStore) User(accessToken string) (interface{}, error) {
	if accessToken != s.accessToken {
		return nil, errors.New("not found")
	}
	return s.user, nil
}

func (s *userStore) OAuth2Token(user interface{}) (*oauth2.Token, error) {
	if !reflect.DeepEqual(user, s.user) {
		return nil, errors.New("unknown user")
	}
	return &oauth2.Token{
		AccessToken:  "xxx",
		RefreshToken: "rrr",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}, nil
}

func (s *userStore) UpdateOAuth2Token(user interface{}, token *oauth2.Token) error {
	if !reflect.DeepEqual(user, s.user) {
		return errors.New("unknown user")
	}
	s.tokenRefreshed = true
	return nil
}

func (s *userStore) Principal(user interface{}) interface{} {
	return user
}

type permMapper struct{}

func (permMapper) RequiredPermissioin(method, path string) (interface{}, error) {
	return method + ":" + path, nil
}

type tokenValidator struct {
	allowed map[string]bool
}

func (v tokenValidator) Validate(token string, perm interface{}) error {
	if b, ok := v.allowed[perm.(string)+":"+token]; !ok || !b {
		return errors.New("forbidden")
	}
	return nil
}

type tokenSourceProvider struct{}

func (tokenSourceProvider) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	return tokenSource{t}
}

type tokenSource struct {
	t *oauth2.Token
}

func (s tokenSource) Token() (*oauth2.Token, error) {
	if s.t.Valid() {
		return s.t, nil
	}
	return &oauth2.Token{
		AccessToken:  "xxx",
		RefreshToken: "rrr",
	}, nil
}
