package route

import "github.com/labstack/echo"

const confirmPasswordURL = "/password-confirm"

func initStatic(e *echo.Echo) {
	e.File("/", "../ui-web/html/index.html")
	e.File("/login", "../ui-web/html/login.html")
	e.File(confirmPasswordURL, "../ui-web/html/password-confirm.html")
	e.Static("/dist", "../ui-web/dist")
}
