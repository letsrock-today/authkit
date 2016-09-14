package route

import "github.com/labstack/echo"

func Init(e *echo.Echo) {
	initReverseProxy(e)
	initStatic(e)
	initAPI(e)
}
