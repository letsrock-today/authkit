package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/mocks"
)

func TestLogout(t *testing.T) {
	assert := assert.New(t)

	as := new(mocks.AuthService)
	as.On(
		"RevokeAccessToken",
		"xxx-access-token").Return(nil)
	us := new(mocks.UserService)
	us.On(
		"RevokeAccessToken",
		"some_provider_id",
		"xxx-access-token").Return(nil)

	e := echo.New()
	req, err := http.NewRequest(echo.GET, "", nil)
	assert.NoError(err)
	req.Header.Set(echo.HeaderAuthorization, "bearer xxx-access-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(rec, e.Logger()))

	h := handler{
		auth:  as,
		users: us,
		config: authkit.Config{
			PrivateOAuth2Provider: authkit.OAuth2Provider{
				ID: "some_provider_id",
			},
		},
	}

	err = h.Logout(c)
	assert.NoError(err)
	assert.Equal(http.StatusOK, rec.Code)
}
