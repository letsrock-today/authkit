package handler

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/persisttoken"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/profile"
)

type Handler interface {
	Profile(c echo.Context) error
	ProfileSave(c echo.Context) error
	Friends(c echo.Context) error
}

func New(
	c authkit.Config,
	ec authkit.ErrorCustomizer,
	us authkit.HandlerUserService,
	ps profile.Service,
	cc authkit.ContextCreator) Handler {
	if ec == nil || ps == nil {
		panic("invalid argument")
	}
	if cc == nil {
		cc = authkit.DefaultContextCreator{}
	}
	return handler{c, ec, us, ps, cc}
}

type handler struct {
	config          authkit.Config
	errorCustomizer authkit.ErrorCustomizer
	us              authkit.HandlerUserService
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
		userTokenStore{h.us, u}).Client(ctx, nil)
}
