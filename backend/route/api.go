package route

import (
	"strings"

	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/handler"
	"github.com/letsrock-today/hydra-sample/backend/middleware"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

func initAPI(e *echo.Echo, ua userapi.UserAPI) {
	e.GET("/api/auth-providers", handler.AuthProviders)
	e.GET("/api/auth-code-urls", handler.AuthCodeURLs)

	cb := func(method, uri string) (scopes []string, resource, action string) {
		hasProfilePrefix := strings.HasPrefix(uri, "/api/profile")
		switch {
		case strings.EqualFold(method, "GET") && hasProfilePrefix:
			return []string{"test.view_profile"}, "rn:api:profile", "get"
		case strings.EqualFold(method, "POST") && hasProfilePrefix:
			return []string{"test.edit_profile"}, "rn:api:profile", "edit"
		}
		return []string{""}, "rn:api", "get"
	}

	m := middleware.AccessToken(config.PrivPID, ua, cb)

	e.GET("/api/profile", handler.Profile, m)
	e.POST("/api/profile", handler.ProfileSave, m)

	e.POST("/api/login", handler.Login)
	e.POST("/api/login-priv", handler.LoginPriv)

	e.GET("/callback", handler.Callback)

	e.POST("/password-reset", handler.ResetPassword)
	e.POST("/password-change", handler.ChangePassword)

	e.GET("/email-confirm", handler.EmailConfirm)
}
