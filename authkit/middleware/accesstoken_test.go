package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/letsrock-today/hydra-sample/authkit"
)

func TestConfigValidation(t *testing.T) {
	assert := assert.New(t)

	us := new(testUserService)

	// No config at all
	// No obligatory settings
	assert.Panics(func() {
		AccessTokenWithConfig(AccessTokenConfig{})
	})

	// Ensure that defaults are used if non-obligatory settings are not set
	var effcfg *AccessTokenConfig
	reportEffectiveConfig = func(c AccessTokenConfig) {
		effcfg = &c
	}
	AccessTokenWithConfig(AccessTokenConfig{
		UserService:    us,
		TokenValidator: testTokenValidator{},
	})
	assert.NotNil(effcfg)
	assert.Equal(DefaultContextKey, effcfg.ContextKey)
	assert.Equal(DefaultPermissionMapper{}, effcfg.PermissionMapper)

	// Ensure that defaults are used in AccessToken except fields set explicitly
	effcfg = nil
	AccessToken(
		"xxx-provider",
		us,
		testTokenValidator{})
	assert.NotNil(effcfg)
	assert.Equal(DefaultContextKey, effcfg.ContextKey)
	assert.Equal(DefaultPermissionMapper{}, effcfg.PermissionMapper)
	assert.NotNil(effcfg.UserService)
	assert.NotNil(effcfg.TokenValidator)
}

func TestAccessTokenWithConfig(t *testing.T) {
	us := new(testUserService)
	us.On(
		"UserByAccessToken",
		"xxx-provider",
		"xxx").Return(testUser{"valid@login.ok", "name"}, nil)
	us.On(
		"UserByAccessToken",
		"xxx-provider",
		"zzz").Return(testUser{"valid@login.ok", "name"}, nil)
	us.On(
		"UserByAccessToken",
		"xxx-provider",
		"yyy").Return(nil, authkit.NewUserNotFoundError(nil))
	us.On(
		"OAuth2Token",
		"valid@login.ok",
		"xxx-provider").Return(&oauth2.Token{
		AccessToken:  "xxx",
		RefreshToken: "rrr",
	}, nil)
	us.On(
		"OAuth2Token",
		"unknown@login.ok",
		"xxx-provider").Return(nil, authkit.NewUserNotFoundError(nil))

	accessTokenConfig := AccessTokenConfig{
		PrivateProviderID: "xxx-provider",
		UserService:       us,
		PermissionMapper:  testPermMapper{},
		TokenValidator: testTokenValidator{
			allowed: map[string]bool{
				"GET:/permitted:xxx":     true,
				"GET:/permitted:yyy":     true,
				"POST:/permitted:xxx":    true,
				"POST:/permitted:yyy":    true,
				"GET:/permitted-get:xxx": true,
				"GET:/permitted-get:yyy": true,
			},
		},
		OAuth2Config:   testTokenSourceProvider{},
		ContextCreator: authkit.DefaultContextCreator{},
	}

	invalidHeaderFormatMsg := errInvalidAuthHeader.Error()
	notPermittedMsg := errAccessDenied.Error()

	cases := []struct {
		name              string
		w                 *httptest.ResponseRecorder
		r                 *http.Request
		accessTokenConfig AccessTokenConfig
		accessTokenHeader string
		expStatusCode     int
		expRespBody       string
		unprotected       bool
	}{
		{
			name:              "Unprotected resource, header empty",
			w:                 httptest.NewRecorder(),
			r:                 testNewGetUnprotected(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "",
			expStatusCode:     http.StatusOK,
			expRespBody:       testNextHandlerMsg,
			unprotected:       true,
		},
		{
			name:              "Protected resource, header empty",
			w:                 httptest.NewRecorder(),
			r:                 testNewGetRestricted(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "",
			expStatusCode:     http.StatusForbidden,
			expRespBody:       invalidHeaderFormatMsg,
		},
		{
			name:              "Protected resource, invalid header format (no 'bearer')",
			w:                 httptest.NewRecorder(),
			r:                 testNewGetRestricted(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "zzz",
			expStatusCode:     http.StatusForbidden,
			expRespBody:       invalidHeaderFormatMsg,
		},
		{
			name:              "Protected resource, invalid header format (no token after 'bearer')",
			w:                 httptest.NewRecorder(),
			r:                 testNewGetRestricted(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "bearer   ",
			expStatusCode:     http.StatusForbidden,
			expRespBody:       invalidHeaderFormatMsg,
		},
		{
			name:              "Restricted resource, valid token",
			w:                 httptest.NewRecorder(),
			r:                 testNewGetRestricted(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "bearer xxx",
			expStatusCode:     http.StatusForbidden,
			expRespBody:       notPermittedMsg,
		},
		{
			name:              "Permitted resource, valid token, GET",
			w:                 httptest.NewRecorder(),
			r:                 testNewGetPermitted(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "bearer xxx",
			expStatusCode:     http.StatusOK,
			expRespBody:       testNextHandlerMsg,
		},
		{
			name:              "Permitted resource, valid token, POST",
			w:                 httptest.NewRecorder(),
			r:                 testNewPostPermitted(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "bearer xxx",
			expStatusCode:     http.StatusOK,
			expRespBody:       testNextHandlerMsg,
		},
		{
			name:              "Permitted resource, invalid token",
			w:                 httptest.NewRecorder(),
			r:                 testNewGetPermitted(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "bearer zzz",
			expStatusCode:     http.StatusForbidden,
			expRespBody:       notPermittedMsg,
		},
		{
			name:              "Permitted resource, unknown token",
			w:                 httptest.NewRecorder(),
			r:                 testNewGetPermitted(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "bearer yyy", // token with permission, but without user
			expStatusCode:     http.StatusForbidden,
			expRespBody:       notPermittedMsg,
		},
		{
			name:              "Permitted only get resource, valid token, GET",
			w:                 httptest.NewRecorder(),
			r:                 testNewGetPermittedOnlyGet(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "bearer xxx",
			expStatusCode:     http.StatusOK,
			expRespBody:       testNextHandlerMsg,
		},
		{
			name:              "Permitted only get resource, valid token, POST",
			w:                 httptest.NewRecorder(),
			r:                 testNewPostPermittedOnlyGet(t),
			accessTokenConfig: accessTokenConfig,
			accessTokenHeader: "bearer xxx",
			expStatusCode:     http.StatusForbidden,
			expRespBody:       notPermittedMsg,
		},
	}

	for _, cs := range cases {
		cs := cs
		// e.Any(...) and brothers should not be used in parallel
		next := testNextHandler{
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
			assert.Equal(cs.expStatusCode, w.Code)
			assert.Equal(cs.expRespBody, string(w.Body.Bytes()))
		})
	}
}

func testNewGetUnprotected(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.GET, "/unprotected", nil)
	assert.NoError(t, err)
	return r
}

func testNewGetPermitted(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.GET, "/permitted", nil)
	assert.NoError(t, err)
	return r
}

func testNewPostPermitted(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.POST, "/permitted", nil)
	assert.NoError(t, err)
	return r
}

func testNewGetPermittedOnlyGet(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.GET, "/permitted-get", nil)
	assert.NoError(t, err)
	return r
}

func testNewPostPermittedOnlyGet(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.POST, "/permitted-get", nil)
	assert.NoError(t, err)
	return r
}

func testNewGetRestricted(t *testing.T) *http.Request {
	r, err := http.NewRequest(echo.GET, "/restricted", nil)
	assert.NoError(t, err)
	return r
}

const testNextHandlerMsg = "Result from next handler"

type testNextHandler struct {
	hasRun         bool
	checkPrincipal bool
}

func (n *testNextHandler) next(c echo.Context) error {
	n.hasRun = true
	// check if user data is available in context
	if !n.checkPrincipal {
		return c.String(http.StatusOK, testNextHandlerMsg)
	}
	u := c.Get(DefaultContextKey)
	user, ok := u.(testUser)
	if !ok {
		return errors.New("no user in context")
	}
	if user.Name != "name" {
		return errors.New("invalid user in context")
	}
	return c.String(http.StatusOK, testNextHandlerMsg)
}

type testUser struct {
	login string
	Name  string
}

func (u testUser) Login() string {
	return u.login
}

func (u testUser) Email() string {
	return u.login
}

func (u testUser) PasswordHash() string {
	return "some-hash"
}

type testUserService struct {
	mock.Mock
}

func (s *testUserService) UserByAccessToken(
	providerID, accessToken string) (authkit.User, authkit.UserServiceError) {
	args := s.Called(providerID, accessToken)
	err := args.Error(1)
	if err != nil {
		return nil, err.(authkit.UserServiceError)
	}
	return args.Get(0).(authkit.User), nil
}

func (s *testUserService) OAuth2Token(
	login, providerID string) (*oauth2.Token, authkit.UserServiceError) {
	args := s.Called(login, providerID)
	err := args.Error(1)
	if err != nil {
		return nil, err.(authkit.UserServiceError)
	}
	return args.Get(0).(*oauth2.Token), nil
}

func (s *testUserService) UpdateOAuth2Token(
	login, providerID string,
	token *oauth2.Token) authkit.UserServiceError {
	args := s.Called(login, providerID, token)
	return args.Error(0).(authkit.UserServiceError)
}

func (s *testUserService) Principal(user authkit.User) interface{} {
	return user
}

type testPermMapper struct{}

func (testPermMapper) RequiredPermissioin(
	method, path string) (interface{}, error) {
	return method + ":" + path, nil
}

type testTokenValidator struct {
	allowed map[string]bool
}

func (v testTokenValidator) Validate(token string, perm interface{}) error {
	if b, ok := v.allowed[perm.(string)+":"+token]; !ok || !b {
		return errors.New("forbidden")
	}
	return nil
}

type testTokenSourceProvider struct{}

func (testTokenSourceProvider) TokenSource(
	ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	return testTokenSource{t}
}

type testTokenSource struct {
	t *oauth2.Token
}

func (s testTokenSource) Token() (*oauth2.Token, error) {
	if s.t.Valid() {
		return s.t, nil
	}
	return &oauth2.Token{
		AccessToken:  "xxx",
		RefreshToken: "rrr",
	}, nil
}
