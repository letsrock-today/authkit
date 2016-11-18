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

func New(
	ac authkit.Config,
	hc authkithandler.Config) Handler {
	if !hc.Valid() {
		panic("invalid argument")
	}
	if hc.ContextCreator == nil {
		hc.ContextCreator = authkit.DefaultContextCreator{}
	}
	return handler{
		ac,
		hc.ErrorCustomizer,
		hc.UserService,
		hc.ProfileService.(profile.Service),
		hc.ContextCreator}
}

type handler struct {
	config          authkit.Config
	errorCustomizer authkit.ErrorCustomizer
	users           authkit.HandlerUserService
	profiles        profile.Service
	contextCreator  authkit.ContextCreator
}

// createHTTPClient creates http.Client via wrapper around oauth2.Config,
// which persists oauth2 tokens in the user store.
func (h handler) createHTTPClient(
	u authkit.User,
	p authkit.OAuth2Provider) *http.Client {
	ctx := h.contextCreator.CreateContext(p.ID)
	return persisttoken.WrapOAuth2Config(
		p.OAuth2Config,
		u.Login(),
		p.ID,
		userTokenStore{h.users, u},
		nil).Client(ctx, nil)
}
