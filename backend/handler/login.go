package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/service/hydra"
)

type (
	loginForm struct {
		Challenge []string `form:"challenge" valid:"required"`
		Login     []string `form:"login" valid:"email,required"`
		Password  []string `form:"password" valid:"stringlength(3|10),required"`
		Scopes    []string `form:"scopes" valid:"stringlength(1|500),required"`
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

	if err := UserService.Authenticate(
		lf.Login[0],
		lf.Password[0]); err != nil {
		return c.JSON(http.StatusOK, newJsonError(err))
	}

	signedTokenString, err := hydra.GenerateConsentToken(
		lf.Login[0],
		lf.Scopes,
		lf.Challenge[0])
	if err != nil {
		return err
	}

	reply := loginReply{
		Consent: signedTokenString,
	}
	return c.JSON(http.StatusOK, reply)
}
