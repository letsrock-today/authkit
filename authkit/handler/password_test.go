package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/mocks"
)

func TestRestorePassword(t *testing.T) {
	us := new(mocks.UserService)
	us.On(
		"User",
		"valid-login").Return(testUser{
		login:        "valid-login",
		passwordHash: "valid_password_hash",
	}, nil)
	us.On(
		"User",
		"unreachable-login").Return(testUser{
		login:        "unreachable-login",
		passwordHash: "valid_password_hash",
	}, nil)
	us.On(
		"User",
		"unknown-login").Return(nil, authkit.NewUserNotFoundError(nil))
	us.On(
		"RequestPasswordChangeConfirmation",
		"valid-login",
		"valid@login.ok",
		"",
		"valid_password_hash").Return(nil)
	us.On(
		"RequestPasswordChangeConfirmation",
		"unreachable-login",
		"unreachable@login.ok",
		"",
		"valid_password_hash").Return(authkit.NewRequestConfirmationError(nil))

	ps := new(mocks.ProfileService)
	ps.On(
		"ConfirmedEmail",
		"unreachable-login").Return("unreachable@login.ok", "", nil)
	ps.On(
		"ConfirmedEmail",
		"valid-login").Return("valid@login.ok", "", nil)

	h := handler{Config{
		ErrorCustomizer: testErrorCustomizer{},
		UserService:     us,
		ProfileService:  ps,
	}}

	govalidator.TagMap["login"] = govalidator.Validator(emailOrLoginValidator)

	cases := []struct {
		name          string
		params        url.Values
		expStatusCode int
		expBody       string
	}{
		{
			name:          "No params",
			params:        make(url.Values),
			expStatusCode: http.StatusBadRequest,
			expBody:       `{"Code":"invalid req param"}`,
		},
		{
			name: "invalid login",
			params: url.Values{
				"login": []string{"@invalidemail"},
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       `{"Code":"invalid req param"}`,
		},
		{
			name: "unknown login",
			params: url.Values{
				"login": []string{"unknown-login"},
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"user auth err"}`,
		},
		{
			name: "valid login",
			params: url.Values{
				"login": []string{"valid-login"},
			},
			expStatusCode: http.StatusOK,
			expBody:       `{}`,
		},
		{
			name: "cannot send email",
			params: url.Values{
				"login": []string{"unreachable-login"},
			},
			expStatusCode: http.StatusInternalServerError,
			expBody:       http.StatusText(http.StatusInternalServerError),
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
				ctx := e.NewContext(req, rec)
				ctx.Request().Header.Set(echo.HeaderContentType, enc.contentType)

				err = h.RestorePassword(ctx)
				if enc.invalid {
					e.HTTPErrorHandler(err, ctx)
					assert.Equal(http.StatusBadRequest, rec.Code)
					assert.Equal(`{"Code":"invalid req param"}`, string(rec.Body.Bytes()))
				} else {
					e.HTTPErrorHandler(err, ctx)
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

	h := handler{Config{
		ErrorCustomizer: testErrorCustomizer{},
		UserService:     us,
		OAuth2State: authkit.OAuth2State{
			TokenIssuer:  "zzz",
			TokenSignKey: []byte("xxx"),
			Expiration:   1 * time.Hour,
		},
		AuthCookieName: "xxx-auth-cookie",
	}}

	govalidator.TagMap["password"] = govalidator.Validator(func(p string) bool {
		// simplified password validator for test
		return len(p) > 3
	})

	cases := []struct {
		name          string
		params        url.Values
		expStatusCode int
		expBody       string
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
					h.Config,
					"valid@login.ok",
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
					h.Config,
					"valid@login.ok",
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
					h.Config,
					"unknown@login.ok",
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
					h.Config,
					"valid@login.ok",
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
					h.Config,
					"valid@login.ok",
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
				ctx := e.NewContext(req, rec)
				ctx.Request().Header.Set(echo.HeaderContentType, enc.contentType)

				err = h.ChangePassword(ctx)
				if enc.invalid {
					e.HTTPErrorHandler(err, ctx)
					assert.Equal(http.StatusBadRequest, rec.Code)
					assert.Equal(`{"Code":"invalid req param"}`, string(rec.Body.Bytes()))
				} else {
					e.HTTPErrorHandler(err, ctx)
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
