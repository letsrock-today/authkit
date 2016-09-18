package handler

import (
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
	"github.com/letsrock-today/hydra-sample/backend/util"
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

var (
	emailInvalidErr    = fmt.Errorf("Email invalid or not registered in the app")
	confirmPasswordURL = "/password-confirm"
)

func ResetPassword(c echo.Context) error {
	var rp resetPasswordForm
	if err := c.Bind(&rp); err != nil {
		return err
	}
	if _, err := govalidator.ValidateStruct(rp); err != nil {
		return c.JSON(http.StatusOK, newJsonError(err))
	}

	user, err := UserService.Get(rp.Email)
	if err != nil {
		if err == userapi.AuthErrorUserNotFound {
			return c.JSON(http.StatusOK, newJsonError(err))
		}
		return err
	}
	cfg := config.GetConfig()
	token, err := newStateToken(
		cfg.OAuth2State.TokenSignKey,
		cfg.OAuth2State.TokenIssuer,
		user.Email,
		user.PasswordHash,
		cfg.PasswordResetLinkLifespan)
	if err != nil {
		return err
	}

	confirmPasswordExternalURL := cfg.ExternalBaseURL + confirmPasswordURL
	link := fmt.Sprintf("%s?token=%s", confirmPasswordExternalURL, token)
	text := fmt.Sprintf("Follow this link to change your password: %s\n", link)
	if err = util.SendEmail(user.Email, "Confirm password reset", text); err != nil {
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

	cfg := config.GetConfig()
	claims, err := parseStateToken(
		cfg.OAuth2State.TokenSignKey,
		cfg.OAuth2State.TokenIssuer,
		cp.Token)
	if err != nil {
		if err, ok := err.(jwt.ValidationError); ok {
			if err.Errors&jwt.ValidationErrorExpired == jwt.ValidationErrorExpired {
				return c.JSON(http.StatusOK, newJsonError(err))
			}
		}
		return err
	}

	user, err := UserService.Get(claims.Audience)
	if err != nil {
		if err == userapi.AuthErrorUserNotFound {
			return c.JSON(http.StatusOK, newJsonError(err))
		}
		return err
	}
	if user.PasswordHash != claims.Subject {
		return c.JSON(http.StatusOK, newJsonError(userapi.AuthError))
	}
	err = UserService.UpdatePassword(user.Email, cp.Password)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct{}{})
}
