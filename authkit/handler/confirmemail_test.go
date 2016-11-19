package handler

import (
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/middleware"
	"github.com/letsrock-today/authkit/authkit/mocks"
)

func TestConfirmEmail(t *testing.T) {
	us := new(mocks.UserService)
	us.On(
		"User",
		"valid@login.ok").Return(testUser{}, nil)
	us.On(
		"User",
		"invalid@login.ok").Return(nil, authkit.NewUserNotFoundError(nil))

	ps := new(mocks.ProfileService)
	ps.On(
		"EnsureExists",
		"valid@login.ok").Return(nil)
	ps.On(
		"SetEmailConfirmed",
		"valid@login.ok",
		"valid@login.ok",
		true).Return(nil)
	ps.On(
		"SetEmailConfirmed",
		"invalid@login.ok",
		"invalid@login.ok",
		true).Return(authkit.NewUserNotFoundError(nil))

	h := handler{Config{
		ErrorCustomizer: testErrorCustomizer{},
		UserService:     us,
		ProfileService:  ps,
		OAuth2State: authkit.OAuth2State{
			TokenIssuer:  "zzz",
			TokenSignKey: []byte("xxx"),
			Expiration:   1 * time.Hour,
		},
	}}

	cases := []struct {
		name          string
		params        url.Values
		expStatusCode int
		expBody       string
	}{
		{
			name:   "No params",
			params: make(url.Values),
			// because token shold not be created by human, but by our own app
			expStatusCode: http.StatusInternalServerError,
			expBody:       http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "invalid token",
			params: url.Values{
				"token": []string{"invalid_token"},
			},
			// because token shold not be created by human, but by our own app
			expStatusCode: http.StatusInternalServerError,
			expBody:       http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "expired token",
			params: url.Values{
				"token": testNewEmailTokenString(
					t, h.Config, "valid@login.ok", "valid@login.ok", "", true),
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `Error: user auth err`,
		},
		{
			name: "invalid login (deleted)",
			params: url.Values{
				"token": testNewEmailTokenString(
					t, h.Config, "invalid@login.ok", "invalid@login.ok", ""),
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `Error: user auth err`,
		},
		{
			name: "everything OK",
			params: url.Values{
				"token": testNewEmailTokenString(
					t, h.Config, "valid@login.ok", "valid@login.ok", ""),
			},
			expStatusCode: http.StatusOK,
			expBody:       `OK`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(st *testing.T) {
			st.Parallel()
			assert := assert.New(st)
			e := echo.New()

			e.Renderer = testTemplateRenderer

			req, err := http.NewRequest(echo.GET, "/confirm-email?"+c.params.Encode(), nil)
			assert.NoError(err)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err = h.ConfirmEmail(ctx)
			e.HTTPErrorHandler(err, ctx)
			assert.Equal(c.expStatusCode, rec.Code)
			assert.Equal(c.expBody, string(rec.Body.Bytes()))
		})
	}
}

func TestSendConfirmationEmail(t *testing.T) {
	ps := new(mocks.ProfileService)
	ps.On(
		"Email",
		"valid-login").Return("valid@login.ok", "Kate", nil)
	us := new(mocks.UserService)
	us.On(
		"RequestEmailConfirmation",
		"valid-login",
		"valid@login.ok",
		"Kate").Return(nil)
	h := handler{Config{
		ErrorCustomizer: testErrorCustomizer{},
		UserService:     us,
		ProfileService:  ps,
	}}
	e := echo.New()
	req := new(http.Request)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(middleware.DefaultContextKey, testUser{login: "valid-login"})
	err := h.SendConfirmationEmail(c)
	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(http.StatusOK, rec.Code)
}

type testTemplate struct {
	templates *template.Template
}

func (t *testTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var testTemplateRenderer = &testTemplate{
	templates: template.Must(template.New(authkit.ConfirmEmailTemplateName).Parse(
		`{{if .Code}}Error: {{.Code}}{{else}}OK{{end}}`)),
}
