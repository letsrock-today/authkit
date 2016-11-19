package handler

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/letsrock-today/authkit/authkit"
	authkithandler "github.com/letsrock-today/authkit/authkit/handler"
	"github.com/letsrock-today/authkit/authkit/persisttoken"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/profile"
)

type Handler interface {
	Profile(c echo.Context) error
	ProfileSave(c echo.Context) error
	Friends(c echo.Context) error
}

func New(c authkithandler.Config) Handler {
	if !c.Valid() {
		panic("invalid argument")
	}
	if c.ContextCreator == nil {
		c.ContextCreator = authkit.DefaultContextCreator{}
	}
	return handler{c, c.ProfileService.(profile.Service)}
}

type handler struct {
	authkithandler.Config
	profiles profile.Service
}

// createHTTPClient creates http.Client via wrapper around oauth2.Config,
// which persists oauth2 tokens in the user store.
func (h handler) createHTTPClient(
	u authkit.User,
	p authkit.OAuth2Provider) *http.Client {
	ctx := h.ContextCreator.CreateContext(p.ID)
	return persisttoken.WrapOAuth2Config(
		p.OAuth2Config,
		u.Login(),
		p.ID,
		userTokenStore{h.UserService, u},
		nil).Client(ctx, nil)
}
