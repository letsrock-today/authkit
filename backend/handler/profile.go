package handler

import (
	"net/http"

	"github.com/labstack/echo"
)

func Profile(c echo.Context) error {
	login := c.Get("user-login").(string)
	p, err := Profiles.Profile(login)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, p)
}
