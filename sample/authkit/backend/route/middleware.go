package route

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/letsrock-today/hydra-sample/authkit"
	_middleware "github.com/letsrock-today/hydra-sample/authkit/middleware"
)

var (
	profileMiddleware echo.MiddlewareFunc
	friendsMiddleware echo.MiddlewareFunc
)

func initMiddleware(
	e *echo.Echo,
	c authkit.Config,
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

	privateProviderID := "" //TODO: from config
	oauth2Config := c.PrivateOAuth2Provider().OAuth2Config()

	profileMiddleware = _middleware.AccessTokenWithConfig(
		_middleware.AccessTokenConfig{
			privateProviderID,
			_middleware.DefaultContextKey,
			nil,
			tokenValidator,
			userService,
			oauth2Config,
			contextCreator,
		})
	friendsMiddleware = _middleware.AccessTokenWithConfig(
		_middleware.AccessTokenConfig{
			privateProviderID,
			_middleware.DefaultContextKey,
			nil,
			tokenValidator,
			userService,
			oauth2Config,
			contextCreator,
		})
}
