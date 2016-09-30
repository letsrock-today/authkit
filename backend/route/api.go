package route

import (
	"errors"
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

	cbProfile := func(method, _ string) (scopes []string, resource, action string, err error) {
		switch strings.ToUpper(method) {
		case "GET":
			return []string{"core"}, "rn:api", "get", nil
		case "POST":
			return []string{"test.profile.edit"}, "rn:api:profile", "edit", nil
		}
		return []string{}, "", "", errors.New("access forbidden")
	}
	midProfile := middleware.AccessToken(config.PrivPID, ua, cbProfile)
	e.GET("/api/profile", handler.Profile, midProfile)
	e.POST("/api/profile", handler.ProfileSave, midProfile)

	cbFriends := func(method, uri string) (scopes []string, resource, action string, err error) {
		switch strings.ToUpper(method) {
		case "GET":
			return []string{"test.friends.view"}, "rn:api:friends", "get", nil
		}
		return []string{}, "", "", errors.New("access forbidden")
	}
	midFriends := middleware.AccessToken(config.PrivPID, ua, cbFriends)
	e.GET("/api/friends", handler.Friends, midFriends)

	e.POST("/api/login", handler.Login)
	e.POST("/api/login-priv", handler.LoginPriv)

	e.GET("/callback", handler.Callback)

	e.POST("/password-reset", handler.ResetPassword)
	e.POST("/password-change", handler.ChangePassword)

	e.GET("/email-confirm", handler.EmailConfirm)
}
