package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/letsrock-today/hydra-sample/backend/service/socialprofile"
)

func Profile(c echo.Context) error {
	login := c.Get("user-login").(string)
	p, err := Profiles.Profile(login)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, p)
}

func ProfileSave(c echo.Context) error {
	login := c.Get("user-login").(string)
	p := new(socialprofile.Profile)
	if err := c.Bind(p); err != nil {
		return err
	}
	if _, err := govalidator.ValidateStruct(p); err != nil {
		return c.JSON(http.StatusOK, newJsonError(err))
	}
	if err := Profiles.Save(login, p); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct{}{})
}
