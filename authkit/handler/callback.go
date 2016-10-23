package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
)

type (
	callbackRequest struct {
		Error            string `form:"error"`
		ErrorDescription string `form:"error_description"`
		State            string `form:"state" valid:"required"`
		Code             string `form:"code" valid:"required"`
	}
)

//TODO: render errors as in ConfirmEmail

func (h handler) Callback(c echo.Context) error {
	var cr callbackRequest
	if err := c.Bind(&cr); err != nil {
		return errors.WithStack(err)
	}

	if cr.Error != "" {
		err := fmt.Errorf(
			"OAuth2 flow failed, error: %s, description: %s",
			cr.Error,
			cr.ErrorDescription)
		return errors.WithStack(err)
	}

	// Check required fields in case cr.Error is empty
	if _, err := govalidator.ValidateStruct(cr); err != nil {
		return errors.WithStack(err)
	}

	oauth2State := h.config.OAuth2State()
	state, err := apptoken.ParseStateToken(
		oauth2State.TokenIssuer(),
		cr.State,
		oauth2State.TokenSignKey())
	if err != nil {
		return errors.WithStack(err)
	}

	var oauth2cfg authkit.OAuth2Config
	privateProvider := h.config.PrivateOAuth2Provider()
	privPID := privateProvider.ID()
	ctx, err := h.contextCreator.CreateContext(privPID)
	if err != nil {
		return errors.WithStack(err)
	}

	if state.ProviderID() == privPID {
		oauth2cfg = privateProvider.PrivateOAuth2Config()
	} else {
		p := h.config.OAuth2ProviderByID(state.ProviderID())
		if p == nil {
			err := fmt.Errorf("Unknown provider: %s", state.ProviderID())
			return errors.WithStack(err)
		}
		oauth2cfg = p.OAuth2Config()
	}

	token, err := oauth2cfg.Exchange(ctx, cr.Code)
	if err != nil {
		return errors.WithStack(err)
	}

	if state.ProviderID() == privPID {
		return h.handlePrivateProvider(c, state, token)
	}

	// If pid is external, we need:
	// - ensure, that internal user exists for external one,
	// - copy profile info into our DB if user hasn't exist,
	// - save external token to be able to use external API in the future,
	// - generate our private provider's token for user and return it to client.

	// Make provider-specific call to external provider for user's profile data.
	// Obtain external user id and profile data.
	pa, err := h.socialProfiles.SocialProfileService(state.ProviderID())
	if err != nil {
		return errors.WithStack(err)
	}
	client := oauth2cfg.Client(ctx, token)
	p, err := pa.SocialProfile(client)
	if err != nil {
		return errors.WithStack(err)
	}
	login := p.Login()

	// Check that internal user exists for external user.
	user, err := h.users.User(login)
	if err != nil {
		if err, ok := err.(authkit.UserNotFoundError); !ok || !err.IsUserNotFound() {
			return errors.WithStack(err)
		}
	}

	freshUser := (user == nil)

	if freshUser {
		// If internal user doesn't exist:
		err := h.createInternalUser(c, login, p)
		if err != nil {
			return err
		}
	}

	// Save external provider's token in the users DB.
	pid := state.ProviderID()
	if err := h.users.UpdateOAuth2Token(login, pid, token); err != nil {
		return errors.WithStack(err)
	}

	// Issue new private provider's token for the user.
	privToken, err := h.issuePrivateProvidersToken(c, login, freshUser)
	if err != nil {
		return err
	}

	// Return private provider's token to client end exit
	// (redirect client to / with token in header).
	cookie := createCookie(
		h.config.AuthCookieName(),
		privToken.AccessToken)
	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/")
}

func (h handler) handlePrivateProvider(
	c echo.Context, state apptoken.StateToken, token *oauth2.Token) error {
	if state.Login() == "" {
		return errors.WithStack(errors.New("illegal state, empty login"))
	}
	pid := h.config.PrivateOAuth2Provider().ID()
	if err := h.users.UpdateOAuth2Token(state.Login(), pid, token); err != nil {
		return errors.WithStack(err)
	}
	// our trusted provider, just return access token to client
	cookie := createCookie(
		h.config.AuthCookieName(),
		token.AccessToken)
	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/")
}

func (h handler) createInternalUser(
	c echo.Context,
	login string,
	p authkit.Profile) error {
	// - Create internal user.
	pass, err := makeRandomPassword() // create long random password
	if err != nil {
		return errors.WithStack(err)
	}
	if err := h.users.CreateEnabled(login, pass); err != nil {
		if err != nil {
			return errors.WithStack(err)
		}
	}
	// - Save user's profile from external provider to our profile db.
	if err := h.profiles.Save(p); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (h handler) issuePrivateProvidersToken(
	c echo.Context,
	login string,
	freshUser bool) (*oauth2.Token, error) {
	// Check if we have one in DB first.
	privPID := h.config.PrivateOAuth2Provider().ID()
	var privToken *oauth2.Token
	if !freshUser {
		var err error
		privToken, err = h.users.OAuth2Token(login, privPID)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	// Use func to simplify condition check.
	issueToken := func() (err error) {
		privToken, err = h.auth.IssueToken(login)
		if err != nil {
			return err
		}
		return h.users.UpdateOAuth2Token(login, privPID, privToken)
	}

	if privToken == nil {
		if err := issueToken(); err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		if !privToken.Valid() {
			if err := issueToken(); err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	return privToken, nil
}

func makeRandomPassword() (string, error) {
	const passwordLen = 20
	b := make([]byte, passwordLen)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func createCookie(authCookieName, accessToken string) *echo.Cookie {
	cookie := new(echo.Cookie)
	cookie.SetName(authCookieName)
	cookie.SetValue(accessToken)
	cookie.SetSecure(true)
	return cookie
}
