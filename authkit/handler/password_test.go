package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/mocks"
)

func TestForgotPassword(t *testing.T) {
	us := new(mocks.UserService)
	us.On(
		"User",
		"valid@login.ok").Return(testUser{
		login:        "valid@login.ok",
		email:        "valid@login.ok",
		passwordHash: "valid_password_hash",
	}, nil)
	us.On(
		"User",
		"unreachable@login.ok").Return(testUser{
		login:        "unreachable@login.ok",
		email:        "unreachable@login.ok",
		passwordHash: "valid_password_hash",
	}, nil)
	us.On(
		"User",
		"unknown@login.ok").Return(nil, authkit.NewUserNotFoundError(nil))
	us.On(
		"RequestPasswordChangeConfirmation",
		"valid@login.ok",
		"valid_password_hash").Return(nil)
	us.On(
		"RequestPasswordChangeConfirmation",
		"unreachable@login.ok",
		"valid_password_hash").Return(authkit.NewRequestConfirmationError(nil))

	h := handler{
		errorCustomizer: testErrorCustomizer{},
		users:           us,
	}

	cases := []struct {
		name          string
		params        url.Values
		expStatusCode int
		expBody       string
		internalError bool
	}{
		{
			name:          "No params",
			params:        make(url.Values),
			expStatusCode: http.StatusBadRequest,
			expBody:       `{"Code":"invalid req param"}`,
		},
		{
			name: "invalid email",
			params: url.Values{
				"email": []string{"invalidemail"},
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `{"Code":"invalid req param"}`,
		},
		{
			name: "unknown email",
			params: url.Values{
				"email": []string{"unknown@login.ok"},
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"user auth err"}`,
		},
		{
			name: "valid email",
			params: url.Values{
				"email": []string{"valid@login.ok"},
			},
			expStatusCode: http.StatusOK,
			expBody:       `{}`,
		},
		{
			name: "cannot send email",
			params: url.Values{
				"email": []string{"unreachable@login.ok"},
			},
			internalError: true,
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

				err = h.RestorePassword(ctx)
				if enc.invalid || c.internalError {
					assert.Error(err)
				} else {
					assert.NoError(err)
					assert.Equal(c.expStatusCode, rec.Code)
					assert.Equal(c.expBody, string(rec.Body.Bytes()))
				}
			})
		}
	}
}

func TestChangePassword(t *testing.T) {
	us := new(mocks.UserService)
	us.On(
		"UpdatePassword",
		"unknown@login.ok",
		"valid_password_hash",
		"strong-password").Return(authkit.NewUserNotFoundError(nil))
	us.On(
		"UpdatePassword",
		"valid@login.ok",
		"invalid_password_hash",
		"strong-password").Return(authkit.NewUserNotFoundError(nil))
	us.On(
		"UpdatePassword",
		"valid@login.ok",
		"valid_password_hash",
		"strong-password").Return(nil)

	h := handler{
		errorCustomizer: testErrorCustomizer{},
		users:           us,
		config: authkit.Config{
			OAuth2State: authkit.OAuth2State{
				TokenIssuer:  "zzz",
				TokenSignKey: []byte("xxx"),
				Expiration:   1 * time.Hour,
			},
			AuthCookieName: "xxx-auth-cookie",
		},
	}

	govalidator.TagMap["password"] = govalidator.Validator(func(p string) bool {
		// simplified password validator for test
		return len(p) > 3
	})

	cases := []struct {
		name          string
		params        url.Values
		expStatusCode int
		expBody       string
		internalError bool
	}{
		{
			name:          "No params",
			params:        make(url.Values),
			expStatusCode: http.StatusBadRequest,
			expBody:       `{"Code":"invalid req param"}`,
		},
		{
			name: "weak password",
			params: url.Values{
				"password1": []string{"xx"},
				"token": testNewEmailTokenString(
					t,
					h.config,
					"valid@login.ok",
					"valid_password_hash"),
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `{"Code":"invalid req param"}`,
		},
		{
			name: "invalid password hash",
			params: url.Values{
				"password1": []string{"strong-password"},
				"token": testNewEmailTokenString(
					t,
					h.config,
					"valid@login.ok",
					"invalid_password_hash"),
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"user auth err"}`,
		},
		{
			name: "unknown user",
			params: url.Values{
				"password1": []string{"strong-password"},
				"token": testNewEmailTokenString(
					t,
					h.config,
					"unknown@login.ok",
					"valid_password_hash"),
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"user auth err"}`,
		},
		{
			name: "expired token",
			params: url.Values{
				"password1": []string{"strong-password"},
				"token": testNewEmailTokenString(
					t,
					h.config,
					"valid@login.ok",
					"valid_password_hash",
					true),
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"user auth err"}`,
		},
		{
			name: "valid params",
			params: url.Values{
				"password1": []string{"strong-password"},
				"token": testNewEmailTokenString(
					t,
					h.config,
					"valid@login.ok",
					"valid_password_hash"),
			},
			expStatusCode: http.StatusOK,
			expBody:       `{}`,
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

				err = h.ChangePassword(ctx)
				if enc.invalid || c.internalError {
					assert.Error(err)
				} else {
					assert.NoError(err)
					assert.Equal(c.expStatusCode, rec.Code)
					assert.Equal(c.expBody, string(rec.Body.Bytes()))
				}
			})
		}
	}
}

func TestDefaultPasswordValidator(t *testing.T) {
	assert := assert.New(t)
	v := defaultPasswordValidator
	assert.False(v(""), "empty")
	assert.False(v("zZ1$"), "short")
	assert.False(v(`0000000000zzzzzzzzzzZZZZZZZZZZ1111111111@@@@@@@@@@2222222222`), "too long")
	assert.False(v("ZZZ111@@@"), "no lowercase letters")
	assert.False(v("zzz222###"), "no uppercase letters")
	assert.False(v("zzzZZZ$$$"), "no digits")
	assert.False(v("zzzZZZ111"), "no other symbols")
	assert.True(v("zX42#!"), "good password")
}
