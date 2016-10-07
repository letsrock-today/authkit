package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/socialprofile"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user/userapi"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/util/echo-querybinder"
)

type (
	emailConfirmRequest struct {
		Token string `form:"token" valid:"required"`
	}
)

func EmailConfirm(c echo.Context) error {
	var r emailConfirmRequest
	if err := querybinder.New().Bind(&r, c); err != nil {
		return err
	}
	if _, err := govalidator.ValidateStruct(r); err != nil {
		return err
	}

	cfg := config.Get()
	t, err := apptoken.ParseEmailToken(
		cfg.OAuth2State.TokenIssuer,
		r.Token,
		cfg.OAuth2State.TokenSignKey)
	if err != nil {
		if err, ok := err.(jwt.ValidationError); ok {
			if err.Errors&jwt.ValidationErrorExpired == jwt.ValidationErrorExpired {
				//TODO: format error
				return c.String(http.StatusOK, err.Error())
			}
		}
		return err
	}

	// Create empty profile for new user.
	if err := Profiles.Save(t.Login(), &socialprofile.Profile{Email: t.Login()}); err != nil {
		return err
	}

	if err := Users.Enable(t.Login()); err != nil {
		if err == userapi.AuthErrorUserNotFound {
			//TODO: format error
			return c.String(http.StatusOK, err.Error())
		}
		return err
	}

	//TODO: format text (use template or html page) or redirect to login page
	return c.String(http.StatusOK, "Account confirmed, try to login")
}
