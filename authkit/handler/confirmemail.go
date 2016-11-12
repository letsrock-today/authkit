package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/apptoken"
	"github.com/letsrock-today/authkit/authkit/middleware"
)

type (
	confirmationRequest struct {
		Token []string `form:"token" valid:"required"`
	}
)

func (h handler) ConfirmEmail(c echo.Context) error {
	var r confirmationRequest
	if err := c.Bind(&r); err != nil {
		return errors.WithStack(err)
	}
	if _, err := govalidator.ValidateStruct(r); err != nil {
		return errors.WithStack(err)
	}

	oauth2State := h.config.OAuth2State
	t, err := apptoken.ParseEmailToken(
		oauth2State.TokenIssuer,
		r.Token[0],
		oauth2State.TokenSignKey)
	if err != nil {
		if err, ok := errors.Cause(err).(*jwt.ValidationError); ok {
			if err.Errors&jwt.ValidationErrorExpired == jwt.ValidationErrorExpired {
				c.Logger().Debugf("%+v", errors.WithStack(err))
				return c.Render(
					http.StatusUnauthorized,
					authkit.ConfirmEmailTemplateName,
					h.errorCustomizer.UserAuthenticationError(err))
			}
		}
		return errors.WithStack(err)
	}

	if err := h.profiles.SetEmailConfirmed(t.Login(), t.Email(), true); err != nil {
		if authkit.IsUserNotFound(err) {
			c.Logger().Debugf("%+v", errors.WithStack(err))
			return c.Render(
				http.StatusUnauthorized,
				authkit.ConfirmEmailTemplateName,
				h.errorCustomizer.UserAuthenticationError(err))
		}
		return errors.WithStack(err)
	}

	return c.Render(http.StatusOK, authkit.ConfirmEmailTemplateName, nil)
}

func (h handler) SendConfirmationEmail(c echo.Context) error {
	u := c.Get(middleware.DefaultContextKey).(authkit.User)
	email, name, err := h.profiles.Email(u.Login())
	if err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.Render(
			http.StatusUnauthorized,
			authkit.ConfirmEmailTemplateName,
			h.errorCustomizer.UserAuthenticationError(err))
	}
	if err := h.users.RequestEmailConfirmation(
		u.Login(),
		email,
		name); err != nil {
		return errors.WithStack(err)
	}
	return c.String(http.StatusOK, "")
}
