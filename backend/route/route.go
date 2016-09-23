package route

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/letsrock-today/hydra-sample/backend/config"
)

func Init(e *echo.Echo) {
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
	}))
	e.Use(middleware.Secure())
	e.Use(middleware.CSRF(config.Get().CSRFSecret))
	initReverseProxy(e)
	initStatic(e)
	initAPI(e)
}
