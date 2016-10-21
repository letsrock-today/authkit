package route

import (
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit"
	authkithandler "github.com/letsrock-today/hydra-sample/authkit/handler"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/handler"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/profile"
)

const confirmEmailURL = "/email-confirm"

func initAPI(
	e *echo.Echo,
	c authkit.Config,
	as authkit.HandlerAuthService,
	us authkit.HandlerUserService,
	ps profile.Service,
	sps authkit.SocialProfileServices,
	cc authkit.ContextCreator) {

	ec := handler.NewErrorCustomizer()

	ah := authkithandler.NewHandler(c, ec, as, us, ps, sps, cc)
	e.GET("/api/auth-providers", ah.AuthProviders)
	e.GET("/api/auth-code-urls", ah.AuthCodeURLs)
	e.POST("/api/login", ah.ConsentLogin)
	e.POST("/api/login-priv", ah.Login)
	e.GET(confirmEmailURL, ah.ConfirmEmail)
	e.POST("/password-reset", ah.RestorePassword)
	e.POST("/password-change", ah.ChangePassword)
	e.GET("/callback", ah.Callback)

	h := handler.New(c, ec, ps, cc)
	e.GET("/api/profile", h.Profile, profileMiddleware)
	e.POST("/api/profile", h.ProfileSave, profileMiddleware)
	e.GET("/api/friends", h.Friends, friendsMiddleware)
}
