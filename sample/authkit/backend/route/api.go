package route

import (
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/handler"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	_handler "github.com/letsrock-today/hydra-sample/sample/authkit/backend/handler"
)

func initAPI(e *echo.Echo) {

	c := config.GetCfg()
	h := handler.NewHandler(c)

	e.GET("/api/auth-providers", h.AuthProviders)
	e.GET("/api/auth-code-urls", h.AuthCodeURLs)

	e.GET("/api/profile", _handler.Profile, profileMiddleware)
	e.POST("/api/profile", _handler.ProfileSave, profileMiddleware)
	e.GET("/api/friends", _handler.Friends, friendsMiddleware)

	e.POST("/api/login", _handler.Login)
	e.POST("/api/login-priv", _handler.LoginPriv)

	e.GET("/callback", _handler.Callback)

	e.POST("/password-reset", _handler.ResetPassword)
	e.POST("/password-change", _handler.ChangePassword)

	e.GET("/email-confirm", _handler.EmailConfirm)
}
