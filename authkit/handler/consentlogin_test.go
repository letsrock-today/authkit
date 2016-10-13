package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"
)

func TestConsentLogin(t *testing.T) {

	as := new(testAuthService)
	as.On(
		"GenerateConsentToken",
		"valid@login.ok",
		[]string{"valid_scope"},
		"unknown_challenge").Return("", errors.New("invalid challenge"))
	as.On(
		"GenerateConsentToken",
		"valid@login.ok",
		[]string{"valid_scope"},
		"valid_challenge").Return("valid_token", nil)
	as.On(
		"GenerateConsentToken",
		"new.valid@login.ok",
		[]string{"valid_scope"},
		"valid_challenge").Return("valid_token", nil)
	as.On(
		"GenerateConsentToken",
		"old.valid@login.ok",
		[]string{"valid_scope"},
		"valid_challenge").Return("valid_token", nil)
	as.On(
		"GenerateConsentToken",
		"broken.valid@login.ok",
		[]string{"valid_scope"},
		"valid_challenge").Return("valid_token", nil)

	us := new(testUserService)
	us.On(
		"Authenticate",
		"valid@login.ok",
		"invalid_password").Return(newTestUserNotFoundError())
	us.On(
		"Authenticate",
		"valid@login.ok",
		"valid_password").Return(nil)
	us.On(
		"Create",
		"new.valid@login.ok",
		"valid_password").Return(newTestAccountDisabledError())
	us.On(
		"Create",
		"broken.valid@login.ok",
		"valid_password").Return(newTestAccountDisabledError())
	us.On(
		"Create",
		"old.valid@login.ok",
		"valid_password").Return(newTestDuplicateUserError())
	us.On(
		"RequestEmailConfirmation",
		"new.valid@login.ok").Return(nil)
	us.On(
		"RequestEmailConfirmation",
		"old.valid@login.ok").Return(nil)
	us.On(
		"RequestEmailConfirmation",
		"broken.valid@login.ok").Return(UserServiceError(errors.New("cannot send email")))

	h := handler{
		errorCustomizer: testErrorCustomizer{},
		auth:            as,
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
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"invalid req param"}`,
		},
		{
			name: "Login: Unknown challenge",
			params: url.Values{
				"action":    []string{"login"},
				"challenge": []string{"unknown_challenge"},
				"login":     []string{"valid@login.ok"},
				"password":  []string{"valid_password"},
				"scopes":    []string{"valid_scope"},
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"user auth err"}`,
		},
		{
			name: "Login: invalid password",
			params: url.Values{
				"action":    []string{"login"},
				"challenge": []string{"valid_challenge"},
				"login":     []string{"valid@login.ok"},
				"password":  []string{"invalid_password"},
				"scopes":    []string{"valid_scope"},
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"user auth err"}`,
		},
		{
			name: "Login: OK",
			params: url.Values{
				"action":    []string{"login"},
				"challenge": []string{"valid_challenge"},
				"login":     []string{"valid@login.ok"},
				"password":  []string{"valid_password"},
				"scopes":    []string{"valid_scope"},
			},
			expStatusCode: http.StatusOK,
			expBody:       `{"consent":"valid_token"}`,
		},
		{
			name: "Signup: OK",
			params: url.Values{
				"action":    []string{"signup"},
				"challenge": []string{"valid_challenge"},
				"login":     []string{"new.valid@login.ok"},
				"password":  []string{"valid_password"},
				"scopes":    []string{"valid_scope"},
			},
			// account disabled until user confirms it
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"acc disabled"}`,
		},
		{
			name: "Signup: duplicate",
			params: url.Values{
				"action":    []string{"signup"},
				"challenge": []string{"valid_challenge"},
				"login":     []string{"old.valid@login.ok"},
				"password":  []string{"valid_password"},
				"scopes":    []string{"valid_scope"},
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"dup user"}`,
		},
		{
			name: "Signup: cannot send email",
			params: url.Values{
				"action":    []string{"signup"},
				"challenge": []string{"valid_challenge"},
				"login":     []string{"broken.valid@login.ok"},
				"password":  []string{"valid_password"},
				"scopes":    []string{"valid_scope"},
			},
			internalError: true,
		},
	}

	for _, c := range cases {
		c := c
		for _, enc := range bodyEncoders {
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
				ctx.Request().Header().Set("Content-Type", enc.contentType)

				err = h.ConsentLogin(ctx)
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
