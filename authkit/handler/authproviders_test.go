package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"
)

func TestAuthProviders(t *testing.T) {
	assert := assert.New(t)
	e := echo.New()
	req := new(http.Request)
	rec := httptest.NewRecorder()
	c := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(rec, e.Logger()))

	cfg := &testConfig{
		modTime: time.Now(),
		oauth2Providers: []testOAuth2Provider{
			testOAuth2Provider{
				id:      "aaa",
				name:    "Aaa",
				iconURL: "http://aaa.aa/icon.png",
			},
			testOAuth2Provider{
				id:      "bbb",
				name:    "Bbb",
				iconURL: "http://bbb.bb/icon.png",
			},
		},
	}

	h := NewHandler(cfg)

	err := h.AuthProviders(c)
	assert.NoError(err)
	assert.Equal(http.StatusOK, rec.Code)
	//TODO: check json in body
	//t.Log(string(rec.Body.Bytes()))
}
