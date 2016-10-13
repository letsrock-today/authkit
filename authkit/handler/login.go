package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
)

type (
	//TODO: remove assamption login == email?

	loginForm struct {
		Action   string `form:"action" valid:"required,matches(login|signup)"`
		Login    string `form:"login" valid:"required,email"`
		Password string `form:"password" valid:"required,stringlength(3|30)"`
	}

	loginReply struct {
		RedirectURL string `json:"redirUrl"`
	}
)

// PrivateLogin is a login handler for "pivate" or "priveleged" client - app's own UI.
func (h handler) Login(c echo.Context) error {
	var lf loginForm
	if err := c.Bind(&lf); err != nil {
		c.Logger().Debug(errors.WithStack(err))
		return err
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		c.Logger().Debug(errors.WithStack(err))
		return c.JSON(
			http.StatusUnauthorized,
			h.errorCustomizer.InvalidRequestParameterError(err))
	}

	var (
		action          func(login, password string) UserServiceError
		errorCustomizer func(error) interface{}
	)

	signup := lf.Action == "signup"

	if signup {
		action = h.users.Create
		errorCustomizer = h.errorCustomizer.UserCreationError
	} else {
		action = h.users.Authenticate
		errorCustomizer = h.errorCustomizer.UserAuthenticationError
	}

	if err := action(lf.Login, lf.Password); err != nil {
		if signup {
			if err, ok := err.(AccountDisabledError); ok && err.IsAccountDisabled() {
				if err := h.users.RequestEmailConfirmation(lf.Login); err != nil {
					c.Logger().Error(errors.WithStack(err))
					return err
				}
			}
		}
		c.Logger().Debug(errors.WithStack(err))
		return c.JSON(http.StatusUnauthorized, errorCustomizer(err))
	}

	pp := h.config.PrivateOAuth2Provider()
	t, err := h.auth.IssueConsentToken(pp.ClientID(), pp.Scopes())
	if err != nil {
		c.Logger().Debug(errors.WithStack(err))
		return err
	}

	state, err := apptoken.NewStateWithLoginTokenString(
		h.config.OAuth2State().TokenIssuer(),
		pp.ID(),
		lf.Login,
		h.config.OAuth2State().Expiration(),
		h.config.OAuth2State().TokenSignKey())
	if err != nil {
		c.Logger().Debug(errors.WithStack(err))
		return err
	}

	u, err := url.Parse(h.config.PrivateOAuth2Config().Endpoint.AuthURL)
	if err != nil {
		c.Logger().Debug(errors.WithStack(err))
		return err
	}
	v := u.Query()
	v.Set("client_id", pp.ClientID())
	v.Set("response_type", "code")
	v.Set("scope", strings.Join(pp.Scopes(), " "))
	v.Set("state", state)
	v.Set("consent", t)
	u.RawQuery = v.Encode()

	reply := loginReply{
		RedirectURL: u.String(),
	}
	c.Logger().Debug(errors.WithStack(err))
	return c.JSON(http.StatusOK, reply)
}
