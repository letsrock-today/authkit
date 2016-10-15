package handler

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
)

type (
	confirmationRequest struct {
		Token []string `form:"token" valid:"required"`
	}
)

func (h handler) ConfirmEmail(c echo.Context) error {
	var r confirmationRequest
	if err := c.Bind(&r); err != nil {
		c.Logger().Debug(errors.WithStack(err))
		return err
	}
	if _, err := govalidator.ValidateStruct(r); err != nil {
		c.Logger().Debug(errors.WithStack(err))
		return err
	}

	t, err := apptoken.ParseEmailToken(
		h.config.OAuth2State().TokenIssuer(),
		r.Token[0],
		h.config.OAuth2State().TokenSignKey())
	if err != nil {
		if err, ok := errors.Cause(err).(*jwt.ValidationError); ok {
			if err.Errors&jwt.ValidationErrorExpired == jwt.ValidationErrorExpired {
				c.Logger().Debug(errors.WithStack(err))
				return c.Render(
					http.StatusUnauthorized,
					ConfirmEmailTemplateName,
					h.errorCustomizer.UserAuthenticationError(err))
			}
		}
		c.Logger().Debug(errors.WithStack(err))
		return err
	}

	// Create empty profile for new user.
	ch := make(chan error, 1)
	go func() {
		if err := h.profiles.EnsureExists(t.Login()); err != nil {
			c.Logger().Debug(errors.WithStack(err))
			ch <- err
		}
		close(ch)
	}()

	if err := h.users.Enable(t.Login()); err != nil {
		<-ch
		if err, ok := err.(UserNotFoundError); ok && err.IsUserNotFound() {
			c.Logger().Debug(errors.WithStack(err))
			return c.Render(
				http.StatusUnauthorized,
				ConfirmEmailTemplateName,
				h.errorCustomizer.UserAuthenticationError(err))
		}
		c.Logger().Debug(errors.WithStack(err))
		return err
	}

	if err := <-ch; err != nil {
		return err
	}

	return c.Render(http.StatusOK, ConfirmEmailTemplateName, nil)
}
