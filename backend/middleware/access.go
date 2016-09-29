package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/letsrock-today/hydra-sample/backend/service/hydra"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

type callbackFunc func(method, uri string) (scopes []string, resource, action string)

type (
	AccessTokenConfig struct {
		// Context key to store user login into context.
		// Optional. Default value "user-login".
		ContextKey string

		// Provider id to fetch token from UserAPI
		PID string

		// UserAPI to get user by token
		UserAPI userapi.UserAPI

		// Callback used to map method and uri to scopes, resource name and action
		Callback callbackFunc
	}
)

var (
	DefaultAccessTokenConfig = AccessTokenConfig{
		ContextKey: "user-login",
	}
)

func AccessToken(
	pid string,
	ua userapi.UserAPI,
	cb callbackFunc) echo.MiddlewareFunc {
	c := DefaultAccessTokenConfig
	c.PID = pid
	c.UserAPI = ua
	c.Callback = cb
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
	if config.Callback == nil {
		panic("Callback must be provided")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			const prefix = "Bearer "

			// Get access token from header.

			auth := req.Header().Get("Authorization")
			split := strings.SplitN(auth, " ", 2)
			if len(split) != 2 || !strings.EqualFold(split[0], "bearer") {
				return echo.NewHTTPError(http.StatusForbidden, "invalid header format")
			}
			token := split[1]

			scopes, resource, action := config.Callback(req.Method(), req.URI())

			if err := hydra.ValidateAccessTokenPermissions(
				token,
				resource,
				action,
				scopes); err != nil {
				log.Println(err)
				return echo.NewHTTPError(http.StatusForbidden, "invalid access token or operation is not permitted")
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

			return next(c)
		}
	}
}
