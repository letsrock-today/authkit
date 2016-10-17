package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"
)

func TestAuthCodeURLs(t *testing.T) {
	assert := assert.New(t)
	e := echo.New()
	req := new(http.Request)
	rec := httptest.NewRecorder()
	c := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(rec, e.Logger()))

	cfg := &testConfig{
		oauth2State: testOAuth2State{
			tokenIssuer:  "zzz",
			tokenSignKey: []byte("xxx"),
			expiration:   1 * time.Hour,
		},
		oauth2Providers: []testOAuth2Provider{
			{
				id: "aaa",
				oauth2Config: &oauth2.Config{
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
				id: "bbb",
				oauth2Config: &oauth2.Config{
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
