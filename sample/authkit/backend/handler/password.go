package handler

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user/userapi"
)

type (
	resetPasswordForm struct {
		Email string `form:"email" valid:"required,email"`
	}

	changePasswordForm struct {
		Password string `form:"password1" valid:"stringlength(3|10),required"`
		Token    string `form:"token" valid:"required"`
	}
)

var emailInvalidErr = errors.New("Email is invalid or not registered in the app")

func ResetPassword(c echo.Context) error {
	var rp resetPasswordForm
	if err := c.Bind(&rp); err != nil {
		return err
	}
	if _, err := govalidator.ValidateStruct(rp); err != nil {
		return c.JSON(http.StatusOK, newJsonError(err))
	}

	user, err := Users.User(rp.Email)
	if err != nil {
		if err == userapi.AuthErrorUserNotFound {
			return c.JSON(http.StatusOK, newJsonError(err))
		}
		return err
	}

	if err := sendConfirmationEmail(
		user.Email,
		user.PasswordHash,
		confirmPasswordURL,
		true); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct{}{})
}

func ChangePassword(c echo.Context) error {
	var cp changePasswordForm
	if err := c.Bind(&cp); err != nil {
		return err
	}
	if _, err := govalidator.ValidateStruct(cp); err != nil {
		return c.JSON(http.StatusOK, newJsonError(err))
	}

	cfg := config.Get()
	t, err := apptoken.ParseEmailToken(
		cfg.OAuth2State.TokenIssuer,
		cp.Token,
		cfg.OAuth2State.TokenSignKey)
	if err != nil {
		if err, ok := err.(jwt.ValidationError); ok {
			if err.Errors&jwt.ValidationErrorExpired == jwt.ValidationErrorExpired {
				return c.JSON(http.StatusOK, newJsonError(err))
			}
		}
		return err
	}

	user, err := Users.User(t.Login())
	if err != nil {
		if err == userapi.AuthErrorUserNotFound {
			return c.JSON(http.StatusOK, newJsonError(err))
		}
		return err
	}
	if user.PasswordHash != t.PasswordHash() {
		return c.JSON(http.StatusOK, newJsonError(userapi.AuthError))
	}
	err = Users.UpdatePassword(user.Email, cp.Password)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct{}{})
}
