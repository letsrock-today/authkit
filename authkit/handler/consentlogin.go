package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/hydra-sample/authkit"
)

type (
	//TODO: remove assamption login == email?

	consentLoginForm struct {
		P         loginForm
		Challenge string   `form:"challenge" valid:"required"`
		Scopes    []string `form:"scopes" valid:"required,stringlength(1|500)"`
	}

	consentLoginReply struct {
		Consent string `json:"consent"`
	}
)

func (h handler) ConsentLogin(c echo.Context) error {
	var lf consentLoginForm
	if err := c.Bind(&lf); err != nil {
		return errors.WithStack(err)
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(
			http.StatusBadRequest,
			h.errorCustomizer.InvalidRequestParameterError(err))
	}

	signedTokenString, err := h.auth.GenerateConsentToken(
		lf.P.Login,
		lf.Scopes,
		lf.Challenge)
	if err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(
			http.StatusUnauthorized,
			h.errorCustomizer.UserAuthenticationError(err))
	}

	var (
		action          func(login, password string) authkit.UserServiceError
		errorCustomizer func(error) interface{}
	)

	signup := lf.P.Action == "signup"

	if signup {
		action = h.users.Create
		errorCustomizer = h.errorCustomizer.UserCreationError
	} else {
		action = h.users.Authenticate
		errorCustomizer = h.errorCustomizer.UserAuthenticationError
	}

	if err := action(lf.P.Login, lf.P.Password); err != nil {
		if signup {
			if authkit.IsAccountDisabled(err) {
				if err := h.users.RequestEmailConfirmation(lf.P.Login); err != nil {
					return errors.WithStack(err)
				}
			}
		}
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(http.StatusUnauthorized, errorCustomizer(err))
	}

	reply := consentLoginReply{
		Consent: signedTokenString,
	}
	return c.JSON(http.StatusOK, reply)
}
