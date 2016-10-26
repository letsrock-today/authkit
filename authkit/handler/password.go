package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
)

type (
	//TODO: remove assumption login == email?

	restorePasswordForm struct {
		Email string `form:"email" valid:"required~email-required,email~email-format"`
	}

	changePasswordForm struct {
		//TODO: rigorous password rules (digits, different case, etc.)
		Password string `form:"password1" valid:"required~password-required,stringlength(3|30)"`
		Token    string `form:"token" valid:"required"`
	}
)

var errEmailInvalid = errors.New("email is invalid or not registered in the app")

func (h handler) RestorePassword(c echo.Context) error {
	var rp restorePasswordForm
	if err := c.Bind(&rp); err != nil {
		return errors.WithStack(err)
	}
	if _, err := govalidator.ValidateStruct(rp); err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(
			http.StatusBadRequest,
			h.errorCustomizer.InvalidRequestParameterError(err))
	}

	user, err := h.users.User(rp.Email)
	if err != nil {
		if authkit.IsUserNotFound(err) {
			c.Logger().Debugf("%+v", errors.WithStack(err))
			return c.JSON(
				http.StatusUnauthorized,
				h.errorCustomizer.UserAuthenticationError(err))
		}
		return errors.WithStack(err)
	}

	if err := h.users.RequestPasswordChangeConfirmation(
		user.Email(),
		user.PasswordHash()); err != nil {
		return errors.WithStack(err)
	}
	return c.JSON(http.StatusOK, struct{}{})
}

func (h handler) ChangePassword(c echo.Context) error {
	var cp changePasswordForm
	if err := c.Bind(&cp); err != nil {
		return errors.WithStack(err)
	}
	if _, err := govalidator.ValidateStruct(cp); err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(
			http.StatusBadRequest,
			h.errorCustomizer.InvalidRequestParameterError(err))
	}

	s := h.config.OAuth2State()
	t, err := apptoken.ParseEmailToken(
		s.TokenIssuer(),
		cp.Token,
		s.TokenSignKey())
	if err != nil {
		if err, ok := errors.Cause(err).(*jwt.ValidationError); ok {
			c.Logger().Debugf("%+v", errors.WithStack(err))
			return c.JSON(
				http.StatusUnauthorized,
				h.errorCustomizer.UserAuthenticationError(err))
		}
		return errors.WithStack(err)
	}

	if err = h.users.UpdatePassword(
		t.Login(),
		t.PasswordHash(),
		cp.Password); err != nil {
		if authkit.IsUserNotFound(err) {
			c.Logger().Debugf("%+v", errors.WithStack(err))
			return c.JSON(
				http.StatusUnauthorized,
				h.errorCustomizer.UserAuthenticationError(err))
		}
		return errors.WithStack(err)
	}
	return c.JSON(http.StatusOK, struct{}{})
}
