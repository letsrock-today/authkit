package route

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/letsrock-today/hydra-sample/backend/config"
	_middleware "github.com/letsrock-today/hydra-sample/backend/middleware"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

func initMiddleware(e *echo.Echo, ua userapi.UserAPI) *echo.Group {
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
	}))
	e.Use(middleware.Secure())
	e.Use(middleware.CSRF(config.Get().CSRFSecret))
	restricted := e.Group("/api/restricted", _middleware.AccessToken(config.PrivPID, ua))
	return restricted
}
