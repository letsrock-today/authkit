package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
	"github.com/letsrock-today/hydra-sample/backend/util/echo-querybinder"
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
	claims, err := parseStateToken(
		cfg.OAuth2State.TokenSignKey,
		cfg.OAuth2State.TokenIssuer,
		r.Token)
	if err != nil {
		if err, ok := err.(jwt.ValidationError); ok {
			if err.Errors&jwt.ValidationErrorExpired == jwt.ValidationErrorExpired {
				//TODO: format error
				return c.String(http.StatusOK, err.Error())
			}
		}
		return err
	}

	if err := UserService.Enable(claims.Audience); err != nil {
		if err == userapi.AuthErrorUserNotFound {
			//TODO: format error
			return c.String(http.StatusOK, err.Error())
		}
		return err
	}

	//TODO: format text (use template or html page) or redirect to login page
	return c.String(http.StatusOK, "Account confirmed, try to login")
}
