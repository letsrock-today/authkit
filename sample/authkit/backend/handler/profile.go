package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/middleware"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/socialprofile"
)

func (h handler) Profile(c echo.Context) error {
	u := c.Get(middleware.DefaultContextKey).(authkit.User)
	p, err := h.profiles.Profile(u.Login())
	if err != nil {
		return errors.WithStack(err)
	}
	return c.JSON(http.StatusOK, p)
}

func (h handler) ProfileSave(c echo.Context) error {
	u := c.Get(middleware.DefaultContextKey).(authkit.User)
	p := new(socialprofile.Profile)
	if err := c.Bind(p); err != nil {
		return errors.WithStack(err)
	}
	p.Email = u.Login() // we cannot change email, because it used as user's id
	if _, err := govalidator.ValidateStruct(p); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			h.errorCustomizer.InvalidRequestParameterError(err))
	}
	//TODO: preserve fields absent in the html form.
	if err := h.profiles.Save(p); err != nil {
		return errors.WithStack(err)
	}
	// return profile as it saved in store (assume, that store API could modify it)
	pf, err := h.profiles.Profile(u.Login())
	if err != nil {
		return errors.WithStack(err)
	}
	return c.JSON(http.StatusOK, pf)
}
