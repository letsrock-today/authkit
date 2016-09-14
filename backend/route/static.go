package route

import "github.com/labstack/echo"

func initStatic(e *echo.Echo) {
	e.File("/", "../ui-web/html/index.html")
	e.File("/login", "../ui-web/html/login.html")
	e.Static("/dist", "../ui-web/dist")
}
