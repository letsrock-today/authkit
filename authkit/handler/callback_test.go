package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCallback(t *testing.T) {

	exttoken := &oauth2.Token{}
	inttoken := &oauth2.Token{}

	us := new(testUserService)
	us.On(
		"UpdateOAuth2Token",
		"valid@login.ok",
		"private-id",
		inttoken).Return(nil)
	us.On(
		"UpdateOAuth2Token",
		"new.valid@login.ok",
		"private-id",
		inttoken).Return(nil)
	us.On(
		"UpdateOAuth2Token",
		"valid@login.ok",
		"external-id",
		exttoken).Return(nil)
	us.On(
		"UpdateOAuth2Token",
		"new.valid@login.ok",
		"external-id-new",
		exttoken).Return(nil)
	us.On(
		"OAuth2Token",
		"valid@login.ok",
		"private-id").Return(inttoken, nil)
	us.On(
		"User",
		"valid@login.ok").Return(testUser{login: "valid@email.ok"}, nil)
	us.On(
		"User",
		"new.valid@login.ok").Return(nil, testUserServiceError{isUserNotFound: true})

	us.On(
		"CreateEnabled",
		"new.valid@login.ok",
		mock.Anything).Return(nil)

	privCfg := new(testOAuth2Config)
	privCfg.On(
		"Exchange",
		mock.Anything,
		"valid_code").Return(inttoken, nil)

	extCfg := new(testOAuth2Config)
	extCfg.On(
		"Exchange",
		mock.Anything,
		"valid_code").Return(exttoken, nil)

	sp := new(testSocialProfileService)
	sp.On(
		"SocialProfile",
		mock.Anything).Return(&testProfile{login: "valid@login.ok"}, nil)

	spn := new(testSocialProfileService)
	spn.On(
		"SocialProfile",
		mock.Anything).Return(&testProfile{login: "new.valid@login.ok"}, nil)

	sps := new(testSocialProfileServices)
	sps.On(
		"SocialProfileService",
		"external-id").Return(sp, nil)
	sps.On(
		"SocialProfileService",
		"external-id-new").Return(spn, nil)

	as := new(testAuthService)
	as.On(
		"IssueToken",
		mock.Anything,
		"valid@login.ok").Return(inttoken, nil)
	as.On(
		"IssueToken",
		mock.Anything,
		"new.valid@login.ok").Return(inttoken, nil)

	ps := new(testProfileService)
	ps.On(
		"Save",
		mock.Anything).Return(nil)

	h := handler{
		errorCustomizer: testErrorCustomizer{},
		auth:            as,
		users:           us,
		profiles:        ps,
		socialProfiles:  sps,
		config: testConfig{
			oauth2State: testOAuth2State{
				tokenIssuer:  "zzz",
				tokenSignKey: []byte("xxx"),
				expiration:   1 * time.Hour,
			},
			privateOAuth2Provider: testOAuth2Provider{
				id:               "private-id",
				oauth2Config:     privCfg,
				privOAuth2Config: privCfg,
			},
			oauth2Providers: []testOAuth2Provider{
				testOAuth2Provider{
					id:           "external-id",
					oauth2Config: extCfg,
				},
				testOAuth2Provider{
					id:           "external-id-new",
					oauth2Config: extCfg,
				},
			},
		},
		contextCreator: authkit.DefaultContextCreator{},
	}

	cases := []struct {
		name          string
		params        url.Values
		expStatusCode int
		expBody       string
		expCookie     string
		internalError bool
	}{
		{
			name:          "No params",
			params:        make(url.Values),
			internalError: true,
		},
		{
			name: "flow error",
			params: url.Values{
				"error":             []string{"some_error"},
				"error_description": []string{"Error description"},
			},
			internalError: true,
		},
		{
			name: "invalid state",
			params: url.Values{
				"state": []string{"invalid_state"},
				"code":  []string{"valid_code"},
			},
			internalError: true,
		},
		{
			name: "everything OK, private",
			params: url.Values{
				"state": testNewStateTokenString(t, h.config, "private-id", "valid@login.ok"),
				"code":  []string{"valid_code"},
			},
			expStatusCode: http.StatusFound,
			expCookie:     "xxx-auth-cookie=; Secure",
		},
		{
			name: "everything OK, external",
			params: url.Values{
				"state": testNewStateTokenString(t, h.config, "external-id", ""),
				"code":  []string{"valid_code"},
			},
			expStatusCode: http.StatusFound,
			expCookie:     "xxx-auth-cookie=; Secure",
		},
		{
			name: "everything OK, external, new user",
			params: url.Values{
				"state": testNewStateTokenString(t, h.config, "external-id-new", ""),
				"code":  []string{"valid_code"},
			},
			expStatusCode: http.StatusFound,
			expCookie:     "xxx-auth-cookie=; Secure",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(st *testing.T) {
			st.Parallel()
			assert := assert.New(st)
			e := echo.New()

			e.SetRenderer(testTemplateRenderer)

			req, err := http.NewRequest(echo.GET, "/callback?"+c.params.Encode(), nil)
			assert.NoError(err)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(
				standard.NewRequest(req, e.Logger()),
				standard.NewResponse(rec, e.Logger()))

			err = h.Callback(ctx)
			if c.internalError {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(c.expStatusCode, rec.Code)
				assert.Equal(c.expBody, string(rec.Body.Bytes()))
				assert.Equal(c.expCookie, rec.Header().Get(echo.HeaderSetCookie))
			}
		})
	}
}
