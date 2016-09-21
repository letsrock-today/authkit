package handler

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/config"
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

	if pid == privPID {
		oauth2cfg = cfg.HydraOAuth2ConfigInt
		//TODO: provide factory for insecure context and app setting
		//TODO: use real certeficates in PROD and remove transport replacement
		ctx = context.WithValue(
			context.Background(),
			oauth2.HTTPClient,
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
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

	if pid == privPID {
		cookie := new(echo.Cookie)
		//TODO: extract constant
		cookie.SetName("X-App-Auth")
		cookie.SetValue(token.AccessToken)
		cookie.SetSecure(true)
		c.SetCookie(cookie)
		return c.Redirect(http.StatusFound, "/")
	}

	// If pid is of our own provider (hydra), return token to client and exit
	// (redirect client to / with token in header).

	// If pid is external, we need:
	// - ensure, that internal user exists for external one,
	// - copy profile info into our DB if user hasn't exist,
	// - save external token to be able to use external API in the future,
	// - generate our provider's (hydra) token for user and return it to client.

	// Check that internal user exists for external user. Use external user ID
	// to obtain internal user ID.

	// If internal user doesn't exist:

	// - Make provider-specifi call to external provider for user's profile data.

	// - Create internal user.

	// - Save user's profile from external provider to our profile db.

	// Use provider-specific call to exchange short-lived token to long-lived one,
	// if possible (facebook).

	// Save external provider's token in the users DB.

	// Issue new hydra token for the user.

	// Return hydra token to client end exit (redirect client to / with token in header).

	s := fmt.Sprintf("Obtained code=%s and state=%s", cr.Code, cr.State)
	log.Println(s)

	ss := fmt.Sprintf("Claims=%#v", claims)
	log.Println(ss)

	return c.String(http.StatusOK, s)
}
