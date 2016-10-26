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

	oauth2State := h.config.OAuth2State()
	t, err := apptoken.ParseEmailToken(
		oauth2State.TokenIssuer(),
		r.Token[0],
		oauth2State.TokenSignKey())
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

	// Create empty profile for new user.
	ch := make(chan error, 1)
	go func() {
		if err := h.profiles.EnsureExists(t.Login()); err != nil {
			c.Logger().Debugf("%+v", errors.WithStack(err))
			ch <- err
		}
		close(ch)
	}()

	if err := h.users.Enable(t.Login()); err != nil {
		<-ch
		if authkit.IsUserNotFound(err) {
			c.Logger().Debugf("%+v", errors.WithStack(err))
			return c.Render(
				http.StatusUnauthorized,
				authkit.ConfirmEmailTemplateName,
				h.errorCustomizer.UserAuthenticationError(err))
		}
		return errors.WithStack(err)
	}

	if err := <-ch; err != nil {
		return errors.WithStack(err)
	}

	return c.Render(http.StatusOK, authkit.ConfirmEmailTemplateName, nil)
}
