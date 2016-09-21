package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/service/hydra"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

type (
	privLoginForm struct {
		Action   string `form:"action" valid:"required,matches(login|signup)"`
		Login    string `form:"login" valid:"email,required"`
		Password string `form:"password" valid:"stringlength(3|10),required"`
	}
	privLoginReply struct {
		RedirectURL string `json:"redirUrl"`
	}
)

// Login for "priveleged" client - app's own UI
func LoginPriv(c echo.Context) error {

	var lf privLoginForm
	if err := c.Bind(&lf); err != nil {
		return err
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		return c.JSON(http.StatusOK, newJsonError(err))
	}

	var action func(login, password string) error
	signup := lf.Action == "signup"

	if signup {
		action = UserService.Create
	} else {
		action = UserService.Authenticate
	}

	if err := action(
		lf.Login,
		lf.Password); err != nil {
		if signup &&
			(err == userapi.AuthErrorDisabled || err == userapi.AuthErrorDup) {
			if err := sendConfirmationEmail(
				lf.Login,
				"",
				confirmEmailURL,
				false); err != nil {
				return err
			}
		}
		return c.JSON(http.StatusOK, newJsonError(err))
	}

	cfg := config.Get()
	signedTokenString, err := hydra.IssueConsentToken(
		cfg.HydraOAuth2Config.ClientID,
		cfg.HydraOAuth2Config.Scopes)
	if err != nil {
		return c.JSON(http.StatusOK, newJsonError(err))
	}

	state, err := newStateToken(
		cfg.OAuth2State.TokenSignKey,
		cfg.OAuth2State.TokenIssuer,
		lf.Login,
		privPID,
		cfg.OAuth2State.Expiration)
	if err != nil {
		return err
	}

	/*
		nonce := make([]byte, 12)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
	*/

	u, err := url.Parse(cfg.HydraOAuth2Config.Endpoint.AuthURL)
	if err != nil {
		return err
	}
	v := u.Query()
	v.Set("client_id", cfg.HydraOAuth2Config.ClientID)
	v.Set("response_type", "code")
	v.Set("scope", strings.Join(cfg.HydraOAuth2Config.Scopes, " "))
	v.Set("state", state)
	//v.Set("nonce", base64.URLEncoding.EncodeToString(nonce))
	v.Set("consent", signedTokenString)
	u.RawQuery = v.Encode()

	reply := privLoginReply{
		RedirectURL: u.String(),
	}
	return c.JSON(http.StatusOK, reply)
}