package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/oauth2"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/mocks"
)

func TestLogin(t *testing.T) {

	as := new(mocks.AuthService)
	as.On(
		"GenerateConsentTokenPriv",
		"valid@login.ok",
		[]string{"some_scope"},
		"some_client_id").Return("valid_token", nil)
	as.On(
		"GenerateConsentTokenPriv",
		"valid@login.ok",
		[]string{"some_scope"},
		"unknown_client_id").Return("", errors.New("unknown_client"))
	as.On(
		"GenerateConsentTokenPriv",
		"new.valid@login.ok",
		[]string{"some_scope"},
		"some_client_id").Return("valid_token", nil)

	us := new(mocks.UserService)
	us.On(
		"Authenticate",
		"valid@login.ok",
		"invalid_password").Return(authkit.NewUserNotFoundError(nil))
	us.On(
		"Authenticate",
		"valid@login.ok",
		"valid_password").Return(nil)
	us.On(
		"Create",
		"new.valid@login.ok",
		"valid_password").Return(nil)
	us.On(
		"Create",
		"broken.valid@login.ok",
		"valid_password").Return(nil)
	us.On(
		"Create",
		"old.valid@login.ok",
		"valid_password").Return(authkit.NewDuplicateUserError(nil))
	us.On(
		"RequestEmailConfirmation",
		"new.valid@login.ok",
		"new.valid@login.ok",
		"").Return(nil)

	ps := new(mocks.ProfileService)
	ps.On(
		"EnsureExists",
		"new.valid@login.ok",
		"new.valid@login.ok").Return(nil)

	h := handler{
		errorCustomizer: testErrorCustomizer{},
		auth:            as,
		users:           us,
		profiles:        ps,
		config: authkit.Config{
			PrivateOAuth2Provider: authkit.OAuth2Provider{
				ID: "some_id",
				OAuth2Config: &oauth2.Config{
					ClientID: "some_client_id",
					Scopes:   []string{"some_scope"},
				},
			},
		},
	}

	h2 := handler{
		errorCustomizer: testErrorCustomizer{},
		auth:            as,
		users:           us,
		config: authkit.Config{
			PrivateOAuth2Provider: authkit.OAuth2Provider{
				ID: "some_id",
				OAuth2Config: &oauth2.Config{
					ClientID: "unknown_client_id",
					Scopes:   []string{"some_scope"},
				},
			},
		},
	}

	govalidator.TagMap["password"] = govalidator.Validator(func(p string) bool {
		// simplified password validator for test
		return len(p) > 3
	})

	cases := []struct {
		name                         string
		params                       url.Values
		expStatusCode                int
		expBody                      string
		expBodyRegex                 bool
		internalError                bool
		failGenerateConsentTokenPriv bool
	}{
		{
			name:          "No params",
			params:        make(url.Values),
			expStatusCode: http.StatusBadRequest,
			expBody:       `{"Code":"invalid req param"}`,
		},
		{
			name: "Login: short password",
			params: url.Values{
				"action":   []string{"login"},
				"login":    []string{"valid@login.ok"},
				"password": []string{"xx"},
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `{"Code":"invalid req param"}`,
		},
		{
			name: "Login: invalid password",
			params: url.Values{
				"action":   []string{"login"},
				"login":    []string{"valid@login.ok"},
				"password": []string{"invalid_password"},
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"user auth err"}`,
		},
		{
			name: "Login: OK",
			params: url.Values{
				"action":   []string{"login"},
				"login":    []string{"valid@login.ok"},
				"password": []string{"valid_password"},
			},
			expStatusCode: http.StatusOK,
			expBodyRegex:  true,
			expBody:       `\{"redirUrl":".*consent=valid_token.*"\}`,
		},
		{
			name: "Signup: OK",
			params: url.Values{
				"action":   []string{"signup"},
				"login":    []string{"new.valid@login.ok"},
				"password": []string{"valid_password"},
			},
			expStatusCode: http.StatusOK,
			expBodyRegex:  true,
			expBody:       `\{"redirUrl":".*consent=valid_token.*"\}`,
		},
		{
			name: "Signup: duplicate",
			params: url.Values{
				"action":   []string{"signup"},
				"login":    []string{"old.valid@login.ok"},
				"password": []string{"valid_password"},
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"dup user"}`,
		},
		{
			name: "Login: fail GenerateConsentTokenPriv",
			params: url.Values{
				"action":   []string{"login"},
				"login":    []string{"valid@login.ok"},
				"password": []string{"valid_password"},
			},
			internalError:                true,
			failGenerateConsentTokenPriv: true,
		},
	}

	for _, c := range cases {
		c := c
		for _, enc := range testBodyEncoders {
			enc := enc
			t.Run(c.name+", "+enc.name, func(st *testing.T) {
				st.Parallel()
				assert := assert.New(st)
				e := echo.New()

				req, err := http.NewRequest(echo.POST, "", enc.encoder(c.params))
				assert.NoError(err)
				rec := httptest.NewRecorder()
				ctx := e.NewContext(
					standard.NewRequest(req, e.Logger()),
					standard.NewResponse(rec, e.Logger()))
				ctx.Request().Header().Set(echo.HeaderContentType, enc.contentType)

				if c.failGenerateConsentTokenPriv {
					err = h2.Login(ctx)
				} else {
					err = h.Login(ctx)
				}
				if enc.invalid || c.internalError {
					assert.Error(err)
				} else {
					assert.NoError(err)
					assert.Equal(c.expStatusCode, rec.Code)
					if c.expBodyRegex {
						assert.Regexp(c.expBody, string(rec.Body.Bytes()))
					} else {
						assert.Equal(c.expBody, string(rec.Body.Bytes()))
					}
				}
			})
		}
	}
}

func TestSimpleLoginValidator(t *testing.T) {
	assert := assert.New(t)
	v := SimpleLoginValidator
	assert.False(v(""), "empty")
	assert.False(v("zZxx"), "short")
	assert.False(v(`0000000000zzzzzzzzzzZZZZZZZZZZ1111111111@@@@@@@@@@2222222222`), "too long")
	assert.False(v("2ZZZ"), "first char is not letter")
	assert.False(v("zzz222###"), "not allowed chars")
	assert.True(v("iRobot-2_0"), "good login")
}

func TestEmailOrLoginValidator(t *testing.T) {
	assert := assert.New(t)
	v := emailOrLoginValidator
	assert.False(v(""), "empty")
	assert.False(v("zZxx"), "short")
	assert.True(v("iRobot-2_0"), "good login")
	assert.True(v("iRobot@gmail.com"), "good email")
}
