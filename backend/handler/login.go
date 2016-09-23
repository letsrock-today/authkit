package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/service/hydra"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

type (
	loginForm struct {
		Action    string   `form:"action" valid:"required,matches(login|signup)"`
		Challenge string   `form:"challenge" valid:"required"`
		Login     string   `form:"login" valid:"required,email"`
		Password  string   `form:"password" valid:"required,stringlength(3|10)"`
		Scopes    []string `form:"scopes" valid:"required,stringlength(1|500)"`
	}
	loginReply struct {
		Consent string `json:"consent"`
	}
)

func Login(c echo.Context) error {
	var lf loginForm
	if err := c.Bind(&lf); err != nil {
		return err
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		return c.JSON(http.StatusOK, newJsonError(err))
	}

	signedTokenString, err := hydra.GenerateConsentToken(
		lf.Login,
		lf.Scopes,
		lf.Challenge)
	if err != nil {
		return err
	}

	var action func(login, password string) error
	signup := lf.Action == "signup"

	if signup {
		action = Users.Create
	} else {
		action = Users.Authenticate
	}

	if err := action(
		lf.Login,
		lf.Password); err != nil {
		if signup &&
			(err == userapi.AuthErrorDisabled || err == userapi.AuthErrorDup) {
			if err := sendConfirmationEmail(
				lf.Login,
				"",
				confirmEmailURL,
				false); err != nil {
				return err
			}
		}
		return c.JSON(http.StatusOK, newJsonError(err))
	}

	reply := loginReply{
		Consent: signedTokenString,
	}
	return c.JSON(http.StatusOK, reply)
}
