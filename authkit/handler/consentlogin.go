package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type (
	loginForm struct {
		Action    string   `form:"action" valid:"required,matches(login|signup)"`
		Challenge string   `form:"challenge" valid:"required"`
		Login     string   `form:"login" valid:"required,email"`
		Password  string   `form:"password" valid:"required,stringlength(3|30)"`
		Scopes    []string `form:"scopes" valid:"required,stringlength(1|500)"`
	}

	loginReply struct {
		Consent string `json:"consent"`
	}
)

func (h handler) ConsentLogin(c echo.Context) error {
	var lf loginForm
	if err := c.Bind(&lf); err != nil {
		c.Logger().Debug(errors.WithStack(err))
		return err
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		return c.JSON(
			http.StatusUnauthorized,
			h.errorCustomizer.InvalidRequestParameterError(err))
	}

	signedTokenString, err := h.auth.GenerateConsentToken(
		lf.Login,
		lf.Scopes,
		lf.Challenge)
	if err != nil {
		c.Logger().Debug(errors.WithStack(err))
		return c.JSON(
			http.StatusUnauthorized,
			h.errorCustomizer.UserAuthenticationError(err))
	}

	var (
		action          func(login, password string) UserServiceError
		errorCustomizer func(error) interface{}
	)

	signup := lf.Action == "signup"

	//TODO: remove assamption login == email?

	if signup {
		action = h.users.Create
		errorCustomizer = h.errorCustomizer.UserCreationError
	} else {
		action = h.users.Authenticate
		errorCustomizer = h.errorCustomizer.UserAuthenticationError
	}

	if err := action(
		lf.Login,
		lf.Password); err != nil {
		if signup &&
			(err.IsAccountDisabled() || err.IsDuplicateUser()) {
			if err := h.users.RequestEmailConfirmation(lf.Login); err != nil {
				c.Logger().Error(errors.WithStack(err))
				return err
			}
		}
		return c.JSON(http.StatusUnauthorized, errorCustomizer(err))
	}

	reply := loginReply{
		Consent: signedTokenString,
	}
	return c.JSON(http.StatusOK, reply)
}
