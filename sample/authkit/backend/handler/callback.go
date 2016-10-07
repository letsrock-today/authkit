package handler

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/hydra"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/socialprofile"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user/userapi"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/util/echo-querybinder"
)

type (
	callbackRequest struct {
		Error            string `form:"error"`
		ErrorDescription string `form:"error_description"`
		State            string `form:"state" valid:"required"`
		Code             string `form:"code" valid:"required"`
	}
)

const authCookieName = "X-App-Auth"

func Callback(c echo.Context) error {
	var cr callbackRequest
	if err := querybinder.New().Bind(&cr, c); err != nil {
		return err
	}

	if cr.Error != "" {
		return fmt.Errorf("OAuth2 flow failed. Error: %s. Description: %s.", cr.Error, cr.ErrorDescription)
	}

	// Check required fields in case cr.Error is empty
	if _, err := govalidator.ValidateStruct(cr); err != nil {
		return err
	}

	cfg := config.Get()
	state, err := apptoken.ParseStateToken(
		cfg.OAuth2State.TokenIssuer,
		cr.State,
		cfg.OAuth2State.TokenSignKey)
	if err != nil {
		return err
	}

	var oauth2cfg oauth2.Config
	ctx := context.Background()
	hydraCtx := context.WithValue(
		context.Background(),
		oauth2.HTTPClient,
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.TLSInsecureSkipVerify},
			}})

	if state.ProviderID() == config.PrivPID {
		oauth2cfg = cfg.HydraOAuth2ConfigInt
		ctx = hydraCtx
	} else {
		var ok bool
		oauth2cfg, ok = cfg.OAuth2Configs[state.ProviderID()]
		if !ok {
			return fmt.Errorf("Unknown provider: %s", state.ProviderID())
		}
	}

	token, err := oauth2cfg.Exchange(ctx, cr.Code)
	if err != nil {
		return err
	}

	if state.ProviderID() == config.PrivPID {
		if state.Login() == "" {
			return errors.New("illegal state, empty login")
		}
		if err = Users.UpdateToken(state.Login(), config.PrivPID, token); err != nil {
			return err
		}
		// our trusted provider, just return access token to client
		cookie := createCookie(token.AccessToken)
		c.SetCookie(cookie)
		return c.Redirect(http.StatusFound, "/")
	}

	// If pid is external, we need:
	// - ensure, that internal user exists for external one,
	// - copy profile info into our DB if user hasn't exist,
	// - save external token to be able to use external API in the future,
	// - generate our provider's (hydra) token for user and return it to client.

	// Make provider-specific call to external provider for user's profile data.
	// Obtain external user id and profile data.
	pa, err := socialprofile.New(state.ProviderID())
	if err != nil {
		return err
	}
	client := oauth2cfg.Client(ctx, token)
	p, err := pa.Profile(client)
	if err != nil {
		return err
	}

	// Check that internal user exists for external user.
	// We use email as unique id for simplicity.
	// It will only work when we can guarantee, that external provider
	// returns unique and valid email.
	// TODO: In general case we should use synthetic internal id
	// and map external id to it.
	user, err := Users.User(p.Email)
	if err != nil && err != userapi.AuthErrorUserNotFound {
		return err
	}

	if user == nil {
		// If internal user doesn't exist:

		// - Create internal user.
		pass, err := makePassword() // create long random password
		if err != nil {
			return err
		}
		if err := Users.Create(p.Email, pass); err != nil {
			if err != userapi.AuthErrorDisabled {
				return err
			}
			if err := Users.Enable(p.Email); err != nil {
				return err
			}
		}
		// - Save user's profile from external provider to our profile db.
		if err := Profiles.Save(p.Email, p); err != nil {
			return err
		}
	}

	// Use provider-specific call to exchange short-lived token to long-lived one,
	// if possible (facebook).
	//TODO: check if it is relevant, use link below to implement
	//TODO: https://github.com/golang/oauth2/issues/154
	//TODO: actually, returned token expires in about a day, it would be great to exchange it for long-lived one

	// Save external provider's token in the users DB.
	if err = Users.UpdateToken(p.Email, state.ProviderID(), token); err != nil {
		return err
	}

	// Issue new hydra token for the user.
	hydratoken, err := Users.Token(p.Email, config.PrivPID)
	if err != nil {
		return err
	}
	issueToken := func() (err error) {
		hydratoken, err = hydra.IssueToken(hydraCtx, p.Email)
		if err != nil {
			return err
		}
		return Users.UpdateToken(p.Email, config.PrivPID, hydratoken)
	}
	if hydratoken == nil {
		if err = issueToken(); err != nil {
			return err
		}
	} else {
		if !hydratoken.Valid() {
			if err = issueToken(); err != nil {
				return err
			}
		}
	}

	// Return hydra token to client end exit (redirect client to / with token in header).
	cookie := createCookie(hydratoken.AccessToken)
	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/")
}

func makePassword() (string, error) {
	const passwordLen = 20
	b := make([]byte, passwordLen)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func createCookie(accessToken string) *echo.Cookie {
	cookie := new(echo.Cookie)
	cookie.SetName(authCookieName)
	cookie.SetValue(accessToken)
	cookie.SetSecure(true)
	return cookie
}
