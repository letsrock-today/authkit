package route

import "github.com/labstack/echo"

func initStatic(e *echo.Echo) {
	e.File("/", "../ui-web/html/index.html")
	e.File("/login", "../ui-web/html/login.html")
	e.File("/password-reset", "../ui-web/html/password-reset.html")
	e.File("/password-confirm", "../ui-web/html/password-confirm.html")
	e.Static("/dist", "../ui-web/dist")
}
