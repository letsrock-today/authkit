package route

import (
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/handler"
)

func initAPI(e *echo.Echo) {
	e.GET("/api/auth-providers", handler.AuthProviders)
	e.GET("/api/auth-code-urls", handler.AuthCodeURLs)
	e.GET("/api/profile", handler.Profile)

	e.POST("/api/login", handler.Login)
	e.POST("/api/login-priv", handler.LoginPriv)

	e.GET("/callback", handler.Callback)

	e.POST("/password-reset", handler.ResetPassword)
	e.POST("/password-change", handler.ChangePassword)

	e.GET("/email-confirm", handler.EmailConfirm)
}
