package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/labstack/echo"
	"github.com/letsrock-today/authkit/authkit"
	"github.com/stretchr/testify/assert"
)

func TestAuthCodeURLs(t *testing.T) {
	assert := assert.New(t)
	e := echo.New()
	req := new(http.Request)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cfg := authkit.Config{
		OAuth2State: authkit.OAuth2State{
			TokenIssuer:  "zzz",
			TokenSignKey: []byte("xxx"),
			Expiration:   1 * time.Hour,
		},
		OAuth2Providers: []authkit.OAuth2Provider{
			{
				ID: "aaa",
				OAuth2Config: &oauth2.Config{
					ClientID:     "aaa-id",
					ClientSecret: "aaa-secret",
					Scopes:       []string{"111", "222"},
					Endpoint: oauth2.Endpoint{
						TokenURL: "https://aaa.aa/token",
						AuthURL:  "https://aaa.aa/auth",
					},
					RedirectURL: "https://aaa.aa/redirect",
				},
			},
			{
				ID: "bbb",
				OAuth2Config: &oauth2.Config{
					ClientID:     "bbb-id",
					ClientSecret: "bbb-secret",
					Scopes:       []string{"111", "222"},
					Endpoint: oauth2.Endpoint{
						TokenURL: "https://bbb.bb/token",
						AuthURL:  "https://bbb.bb/auth",
					},
					RedirectURL: "https://bbb.bb/redirect",
				},
			},
		},
	}

	h := handler{config: cfg}

	err := h.AuthCodeURLs(c)
	assert.NoError(err)
	assert.Equal(http.StatusOK, rec.Code)
	assert.Regexp(`\{"urls":\[.*"id":"bbb","url":"https://bbb.bb/auth.*\]\}`, string(rec.Body.Bytes()))
}
