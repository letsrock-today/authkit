package handler

import (
	"net/http"
	"net/url"
	"strings"
	"unicode"

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
		Login    string `form:"login" valid:"required~login-required,login~login-format"`
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
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(
			http.StatusBadRequest,
			h.ErrorCustomizer.InvalidRequestParameterError(flatten(err)))
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(
			http.StatusBadRequest,
			h.ErrorCustomizer.InvalidRequestParameterError(err))
	}

	var (
		action          func(login, password string) authkit.UserServiceError
		customizedError func(error) interface{}
		email           = ""
	)

	signup := lf.Action == "signup"

	if signup {
		action = func(login, password string) authkit.UserServiceError {
			if err := h.UserService.Create(login, password); err != nil {
				return err
			}
			// Create empty profile for new user.
			if govalidator.IsEmail(login) {
				email = login
			}
			if err := h.ProfileService.EnsureExists(login, email); err != nil {
				return err
			}
			return nil
		}

		customizedError = h.ErrorCustomizer.UserCreationError
	} else {
		action = h.UserService.Authenticate
		customizedError = h.ErrorCustomizer.UserAuthenticationError
	}

	if err := action(lf.Login, lf.Password); err != nil {
		c.Logger().Debugf("%+v", errors.WithStack(err))
		return c.JSON(http.StatusUnauthorized, customizedError(err))
	}

	pp := h.PrivateOAuth2Provider
	cfg := pp.OAuth2Config.(*oauth2.Config)
	t, err := h.AuthService.GenerateConsentTokenPriv(
		lf.Login,
		cfg.Scopes,
		cfg.ClientID)
	if err != nil {
		return errors.WithStack(err)
	}

	if email != "" {
		go func() {
			if err := h.UserService.RequestEmailConfirmation(
				lf.Login,
				email,
				""); err != nil {
				c.Logger().Debugf("%+v", errors.WithStack(err))
			}
		}()
	}

	s := h.OAuth2State
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

// SimpleLoginValidator checks user name (login) against following rules:
// - length 5..20 chars
// - first char is a letter
// - allowed chars: latin letters, digits, hyphen, underscore.
// Note: these restrictions are pure arbitrary, just to have some
// (we should have some as a proiphylactic against injections).
func SimpleLoginValidator(s string) bool {
	l := len(s)
	if l < 5 || l > 20 {
		return false
	}
	if !unicode.IsLetter([]rune(s)[0]) {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) &&
			!unicode.IsDigit(r) &&
			r != '-' &&
			r != '_' {
			return false
		}
	}
	return true
}

func emailOrLoginValidator(s string) bool {
	return SimpleLoginValidator(s) || govalidator.IsEmail(s)
}
