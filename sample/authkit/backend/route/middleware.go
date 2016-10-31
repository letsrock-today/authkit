package route

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/letsrock-today/authkit/authkit"
	_middleware "github.com/letsrock-today/authkit/authkit/middleware"
	"github.com/letsrock-today/authkit/sample/authkit/backend/config"
)

var (
	profileMiddleware echo.MiddlewareFunc
	friendsMiddleware echo.MiddlewareFunc
)

func initMiddleware(
	e *echo.Echo,
	c config.Config,
	tokenValidator authkit.TokenValidator,
	userService authkit.MiddlewareUserService,
	contextCreator authkit.ContextCreator) {
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
	}))
	e.Use(middleware.Secure())
	e.Use(middleware.CSRF())

	oauth2Config := c.PrivateOAuth2Provider.PrivateOAuth2Config

	profileMiddleware = _middleware.AccessTokenWithConfig(
		_middleware.AccessTokenConfig{
			PrivateProviderID: c.PrivateProviderID,
			ContextKey:        _middleware.DefaultContextKey,
			TokenValidator:    tokenValidator,
			UserService:       userService,
			OAuth2Config:      oauth2Config,
			ContextCreator:    contextCreator,
		})
	friendsMiddleware = _middleware.AccessTokenWithConfig(
		_middleware.AccessTokenConfig{
			PrivateProviderID: c.PrivateProviderID,
			ContextKey:        _middleware.DefaultContextKey,
			TokenValidator:    tokenValidator,
			UserService:       userService,
			OAuth2Config:      oauth2Config,
			ContextCreator:    contextCreator,
		})
}
