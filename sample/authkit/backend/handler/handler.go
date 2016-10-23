package handler

import (
	"github.com/labstack/echo"
	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/profile"
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
	if c == nil || ec == nil || ps == nil {
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
	profiles        profile.Service
	us              authkit.HandlerUserService
	contextCreator  authkit.ContextCreator
}

func (h handler) withOAuthTokenDo(
	u user.User,
	provider authkit.OAuth2Provider,
	do func(client *http.Client) error) error {
	token := u.OAuth2TokenByProviderID(p.ID())
	ctx := h.contextCreator.CreateContext(p.ID())
	client := p.OAuth2Config().Client(ctx, token)
	if err := do(client); err != nil {
		return err
	}
	newToken := config.OAuth2Config.TokenSource(ctx, token).Token()
	if newToken != nil && newToken != token {
		return us.UpdateOAuth2Token(u.Login(), p.ID(), newToken)
	}
	return nil
}
