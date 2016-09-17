package route

import "github.com/labstack/echo"

func initStatic(e *echo.Echo) {
	e.File("/", "../ui-web/html/index.html")
	e.File("/login", "../ui-web/html/login.html")
	e.File("/password-reset", "../ui-web/html/reset-password.html")
	e.Static("/dist", "../ui-web/dist")
}
