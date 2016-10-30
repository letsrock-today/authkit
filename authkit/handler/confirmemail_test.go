package handler

import (
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/mocks"
)

func TestConfirmEmail(t *testing.T) {
	us := new(mocks.UserService)
	us.On(
		"Enable",
		"valid@login.ok").Return(nil)
	us.On(
		"Enable",
		"fail_create_profile@login.ok").Return(nil)
	us.On(
		"Enable",
		"invalid@login.ok").Return(authkit.NewUserNotFoundError(nil))

	ps := new(mocks.ProfileService)
	ps.On(
		"EnsureExists",
		"valid@login.ok").Return(nil)
	ps.On(
		"EnsureExists",
		"invalid@login.ok").Return(nil)
	ps.On(
		"EnsureExists",
		"fail_create_profile@login.ok").Return(errors.New("cannot create profile"))

	h := handler{
		errorCustomizer: testErrorCustomizer{},
		users:           us,
		profiles:        ps,
		config: authkit.Config{
			OAuth2State: authkit.OAuth2State{
				TokenIssuer:  "zzz",
				TokenSignKey: []byte("xxx"),
				Expiration:   1 * time.Hour,
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
			name:   "No params",
			params: make(url.Values),
			// because token shold not be created by human, but by our own app
			internalError: true,
		},
		{
			name: "invalid token",
			params: url.Values{
				"token": []string{"invalid_token"},
			},
			// because token shold not be created by human, but by our own app
			internalError: true,
		},
		{
			name: "expired token",
			params: url.Values{
				"token": testNewEmailTokenString(t, h.config, "valid@login.ok", "", true),
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `Error: user auth err`,
		},
		{
			name: "cannot create profile",
			params: url.Values{
				"token": testNewEmailTokenString(t, h.config, "fail_create_profile@login.ok", ""),
			},
			// it may be other error in PROD if user.Enable() fail first
			internalError: true,
		},
		{
			name: "invalid login (deleted)",
			params: url.Values{
				"token": testNewEmailTokenString(t, h.config, "invalid@login.ok", ""),
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `Error: user auth err`,
		},
		{
			name: "everything OK",
			params: url.Values{
				"token": testNewEmailTokenString(t, h.config, "valid@login.ok", ""),
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

			e.SetRenderer(testTemplateRenderer)

			req, err := http.NewRequest(echo.GET, "/confirm-email?"+c.params.Encode(), nil)
			assert.NoError(err)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(
				standard.NewRequest(req, e.Logger()),
				standard.NewResponse(rec, e.Logger()))

			err = h.ConfirmEmail(ctx)
			if c.internalError {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(c.expStatusCode, rec.Code)
				assert.Equal(c.expBody, string(rec.Body.Bytes()))
			}
		})
	}
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
