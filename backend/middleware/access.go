package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/letsrock-today/hydra-sample/backend/service/hydra"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

type (
	AccessTokenConfig struct {
		// Context key to store user login into context.
		// Optional. Default value "user-login".
		ContextKey string `json:"user_login"`

		// Provider id to fetch token from UserAPI
		PID string `json:"pid"`

		// UserAPI to get user by token
		UserAPI userapi.UserAPI
	}
)

var (
	DefaultAccessTokenConfig = AccessTokenConfig{
		ContextKey: "user-login",
	}
)

func AccessToken(pid string, ua userapi.UserAPI) echo.MiddlewareFunc {
	c := DefaultAccessTokenConfig
	c.PID = pid
	c.UserAPI = ua
	return AccessTokenWithConfig(c)
}

func AccessTokenWithConfig(config AccessTokenConfig) echo.MiddlewareFunc {
	// Defaults
	if config.ContextKey == "" {
		config.ContextKey = DefaultAccessTokenConfig.ContextKey
	}
	if config.PID == "" {
		panic("PID must be provided")
	}
	if config.UserAPI == nil {
		panic("UserAPI must be provided")
	}

	// Initialize
	//TODO

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			const prefix = "Bearer "

			// Get access token from header.

			header := req.Header().Get("Authorization")
			if !strings.HasPrefix(header, prefix) {
				return echo.NewHTTPError(http.StatusForbidden, "invalid header format")
			}
			token := strings.TrimPrefix(header, prefix)

			// Use Hydra to validate access token.
			if err := hydra.ValidateAccessToken(token); err != nil {
				log.Println(err)
				return echo.NewHTTPError(http.StatusForbidden, "invalid token")
			}

			// Get user login from DB by access token.
			user, err := config.UserAPI.UserByToken(config.PID, "accesstoken", token)
			if err != nil {
				log.Println(err)
				if err == userapi.AuthErrorUserNotFound {
					return echo.NewHTTPError(http.StatusForbidden, "invalid token")
				}
				return echo.NewHTTPError(http.StatusInternalServerError)
			}

			// Store user login to context.

			//TODO: provide user.ID and use it here
			c.Set(config.ContextKey, user.Email)

			// TODO: Use Hydra to check token permissions (req.Method(), req.URI()).
			if ok := hydra.CheckAccessTokenPermission(token, req.Method(), req.URI()); !ok {
				return echo.NewHTTPError(http.StatusForbidden, "invalid csrf token")
			}
			return next(c)
		}
	}
}
