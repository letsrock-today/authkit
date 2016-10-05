package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/middleware"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/socialprofile"
)

func Profile(c echo.Context) error {
	login := c.Get(middleware.DefaultContextKey).(string)
	p, err := Profiles.Profile(login)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, p)
}

func ProfileSave(c echo.Context) error {
	login := c.Get(middleware.DefaultContextKey).(string)
	p := new(socialprofile.Profile)
	if err := c.Bind(p); err != nil {
		return err
	}
	p.Email = login // we cannot change email, because it used as user's id
	if _, err := govalidator.ValidateStruct(p); err != nil {
		return c.JSON(http.StatusOK, newJsonError(err))
	}
	//TODO: preserve fields absent in the html form.
	if err := Profiles.Save(login, p); err != nil {
		return err
	}
	// return profile as it saved in store (assume, that store API could modify it)
	if p, err := Profiles.Profile(login); err != nil {
		return err
	} else {
		return c.JSON(http.StatusOK, p)
	}
}
