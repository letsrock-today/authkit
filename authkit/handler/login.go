package handler

import (
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/apptoken"
)

type (
	loginForm struct {
		Action   string `form:"action" valid:"required,matches(login|signup)"`
		Login    string `form:"login" valid:"required~login-required,email~login-format"`
		Password string `form:"password" valid:"required~password-required,password~password-format"`
	}

	loginReply struct {
		RedirectURL string `json:"redirUrl"`
	}
)

// Login is a login handler for "pivate" or "priveleged" client - app's own UI.
func (h handler) Login(c echo.Context) error {
	var lf loginForm
	if err := c.Bind(&lf); err != nil {
		return errors.WithStack(err)
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(
			http.StatusBadRequest,
			h.errorCustomizer.InvalidRequestParameterError(err))
	}

	var (
		action          func(login, password string) authkit.UserServiceError
		customizedError func(error) interface{}
	)

	signup := lf.Action == "signup"

	if signup {
		action = h.users.Create
		customizedError = h.errorCustomizer.UserCreationError
	} else {
		action = h.users.Authenticate
		customizedError = h.errorCustomizer.UserAuthenticationError
	}

	if err := action(lf.Login, lf.Password); err != nil {
		if signup {
			if authkit.IsAccountDisabled(err) {
				if err := h.users.RequestEmailConfirmation(lf.Login); err != nil {
					return errors.WithStack(err)
				}
			}
		}
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(http.StatusUnauthorized, customizedError(err))
	}

	pp := h.config.PrivateOAuth2Provider
	cfg := pp.OAuth2Config.(*oauth2.Config)
	t, err := h.auth.GenerateConsentTokenPriv(
		lf.Login,
		cfg.Scopes,
		cfg.ClientID)
	if err != nil {
		return errors.WithStack(err)
	}

	s := h.config.OAuth2State
	state, err := apptoken.NewStateWithLoginTokenString(
		s.TokenIssuer,
		pp.ID,
		lf.Login,
		s.Expiration,
		s.TokenSignKey)
	if err != nil {
		return errors.WithStack(err)
	}

	u, err := url.Parse(cfg.Endpoint.AuthURL)
	if err != nil {
		return errors.WithStack(err)
	}
	v := u.Query()
	v.Set("client_id", cfg.ClientID)
	v.Set("response_type", "code")
	v.Set("scope", strings.Join(cfg.Scopes, " "))
	v.Set("state", state)
	v.Set("consent", t)
	u.RawQuery = v.Encode()

	reply := loginReply{
		RedirectURL: u.String(),
	}
	c.Logger().Debugf("%+v", errors.WithStack(err))
	return c.JSON(http.StatusOK, reply)
}
