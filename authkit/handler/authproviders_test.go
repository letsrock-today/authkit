package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/letsrock-today/authkit/authkit"
	"github.com/stretchr/testify/assert"
)

func TestAuthProviders(t *testing.T) {
	assert := assert.New(t)
	e := echo.New()
	req := new(http.Request)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cfg := Config{
		ModTime: time.Now(),
		OAuth2Providers: []authkit.OAuth2Provider{
			{
				ID:      "aaa",
				Name:    "Aaa",
				IconURL: "http://aaa.aa/icon.png",
			},
			{
				ID:      "bbb",
				Name:    "Bbb",
				IconURL: "http://bbb.bb/icon.png",
			},
		},
	}

	h := handler{cfg}

	err := h.AuthProviders(c)
	assert.NoError(err)
	assert.Equal(http.StatusOK, rec.Code)
	assert.Regexp(`\{"providers":\[.*\{"id":"bbb","name":"Bbb","iconUrl":"http://bbb.bb/icon.png"\}\].*\}`, string(rec.Body.Bytes()))
}
