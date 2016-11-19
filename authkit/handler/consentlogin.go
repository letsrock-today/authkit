package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/authkit/authkit"
)

type (
	consentLoginForm struct {
		P         loginForm `valid:"required"`
		Challenge string    `form:"challenge" valid:"required"`
		Scopes    []string  `form:"scopes" valid:"required,stringlength(1|500)"`
	}

	consentLoginReply struct {
		Consent string `json:"consent"`
	}
)

func (h handler) ConsentLogin(c echo.Context) error {
	var lf consentLoginForm
	if err := c.Bind(&lf); err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(
			http.StatusBadRequest,
			h.ErrorCustomizer.InvalidRequestParameterError(flatten(err)))
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(
			http.StatusBadRequest,
			h.ErrorCustomizer.InvalidRequestParameterError(flatten(err)))
	}

	signedTokenString, err := h.AuthService.GenerateConsentToken(
		lf.P.Login,
		lf.Scopes,
		lf.Challenge)
	if err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(
			http.StatusUnauthorized,
			h.ErrorCustomizer.UserAuthenticationError(err))
	}

	var (
		action          func(login, password string) authkit.UserServiceError
		errorCustomizer func(error) interface{}
	)

	signup := lf.P.Action == "signup"

	if signup {
		action = h.UserService.Create
		errorCustomizer = h.ErrorCustomizer.UserCreationError
	} else {
		action = h.UserService.Authenticate
		errorCustomizer = h.ErrorCustomizer.UserAuthenticationError
	}

	if err := action(lf.P.Login, lf.P.Password); err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(http.StatusUnauthorized, errorCustomizer(err))
	}

	reply := consentLoginReply{
		Consent: signedTokenString,
	}
	return c.JSON(http.StatusOK, reply)
}

// See https://github.com/asaskevich/govalidator/issues/164
func flatten(err error) error {
	e, ok := err.(govalidator.Errors)
	if !ok {
		return err
	}
	r := govalidator.Errors{}
	for _, v := range e {
		if v, ok := v.(govalidator.Errors); ok {
			f := flatten(v)
			if f, ok := f.(govalidator.Errors); ok {
				r = append(r, f...)
				continue
			}
		}
		r = append(r, v)
	}
	return r
}
