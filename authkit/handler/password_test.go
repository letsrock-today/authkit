package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"
)

func TestForgotPassword(t *testing.T) {
	us := new(testUserService)
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
		"unknown@login.ok").Return(nil, testNewTestUserNotFoundError())
	us.On(
		"RequestPasswordChangeConfirmation",
		"valid@login.ok",
		"valid_password_hash").Return(nil)
	us.On(
		"RequestPasswordChangeConfirmation",
		"unreachable@login.ok",
		"valid_password_hash").Return(testUserServiceError{})

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
	us := new(testUserService)
	us.On(
		"UpdatePassword",
		"unknown@login.ok",
		"valid_password_hash",
		"strong-password").Return(testUserServiceError{isUserNotFound: true})
	us.On(
		"UpdatePassword",
		"valid@login.ok",
		"invalid_password_hash",
		"strong-password").Return(testUserServiceError{isUserNotFound: true})
	us.On(
		"UpdatePassword",
		"valid@login.ok",
		"valid_password_hash",
		"strong-password").Return(nil)

	h := handler{
		errorCustomizer: testErrorCustomizer{},
		users:           us,
		config: testConfig{
			oauth2State: testOAuth2State{
				tokenIssuer:  "zzz",
				tokenSignKey: []byte("xxx"),
				expiration:   1 * time.Hour,
			},
		},
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
