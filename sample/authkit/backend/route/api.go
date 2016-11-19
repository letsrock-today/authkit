package route

import (
	"github.com/labstack/echo"

	authkithandler "github.com/letsrock-today/authkit/authkit/handler"
	"github.com/letsrock-today/authkit/sample/authkit/backend/handler"
)

const confirmEmailURL = "/email-confirm"

func initAPI(
	e *echo.Echo,
	c authkithandler.Config) {

	ah := authkithandler.NewHandler(c)
	e.GET("/api/auth-providers", ah.AuthProviders)
	e.GET("/api/auth-code-urls", ah.AuthCodeURLs)
	e.POST("/api/login", ah.ConsentLogin)
	e.POST("/api/login-priv", ah.Login)
	e.GET("/api/logout", ah.Logout)
	e.GET(confirmEmailURL, ah.ConfirmEmail)
	e.POST("/password-reset", ah.RestorePassword)
	e.POST("/password-change", ah.ChangePassword)
	e.GET("/callback", ah.Callback)

	h := handler.New(c)
	e.POST("/api/confirm-email", ah.SendConfirmationEmail, middlwr)
	e.GET("/api/profile", h.Profile, middlwr)
	e.POST("/api/profile", h.ProfileSave, middlwr)
	e.GET("/api/friends", h.Friends, middlwr)
}
