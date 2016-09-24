package handler

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/service/hydra"
	"github.com/letsrock-today/hydra-sample/backend/service/socialprofile"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
	"github.com/letsrock-today/hydra-sample/backend/util/echo-querybinder"
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
	claims, err := parseStateToken(
		cfg.OAuth2State.TokenSignKey,
		cfg.OAuth2State.TokenIssuer,
		cr.State)
	if err != nil {
		return err
	}

	var oauth2cfg oauth2.Config
	ctx := context.Background()
	pid := claims.Subject

	if pid == config.PrivPID {
		oauth2cfg = cfg.HydraOAuth2ConfigInt
		//TODO: use real certeficates in PROD and remove transport replacement
		ctx = context.WithValue(
			context.Background(),
			oauth2.HTTPClient,
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.TLSInsecureSkipVerify},
				}})
	} else {
		var ok bool
		oauth2cfg, ok = cfg.OAuth2Configs[pid]
		if !ok {
			return fmt.Errorf("Unknown provider: %s", pid)
		}
	}

	token, err := oauth2cfg.Exchange(ctx, cr.Code)
	if err != nil {
		return err
	}

	if pid == config.PrivPID {
		if err = updateToken(claims.Audience, config.PrivPID, token); err != nil {
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
	pa, err := socialprofile.New(pid)
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

		// - Save user's profile from external provider to our profile db.
		if err := Profiles.Save(client, p); err != nil {
			return err
		}

		// - Create internal user.
		pass, err := makePassword() // create long random password
		if err != nil {
			return err
		}
		if err := Users.Create(p.Email, pass); err != nil {
			return err
		}
		if err := Users.Enable(p.Email); err != nil {
			return err
		}
	}

	// Use provider-specific call to exchange short-lived token to long-lived one,
	// if possible (facebook).
	//TODO: check if it is relevant, use link below to implement
	//TODO: https://github.com/golang/oauth2/issues/154

	// Save external provider's token in the users DB.
	if err = updateToken(p.Email, pid, token); err != nil {
		return err
	}

	// Issue new hydra token for the user.
	strtoken, err := Users.Token(p.Email, config.PrivPID)
	if err != nil {
		return err
	}
	var hydratoken *oauth2.Token
	if strtoken == "" {
		hydratoken, err = hydra.IssueToken()
		if err != nil {
			return err
		}
		if err = updateToken(p.Email, config.PrivPID, hydratoken); err != nil {
			return err
		}
	} else {
		// TODO check expiration and update hydra token
		err := json.Unmarshal([]byte(strtoken), hydratoken)
		if err != nil {
			return err
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

func updateToken(login, pid string, token *oauth2.Token) error {
	b, err := json.Marshal(token)
	if err != nil {
		return err
	}
	Users.UpdateToken(login, pid, string(b))
	return err
}
