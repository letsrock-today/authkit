package handler

import (
	"fmt"

	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/profile/profileapi"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user/userapi"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/util/email"
)

var (
	Users    userapi.UserAPI
	Profiles profileapi.ProfileAPI
)

const (
	confirmPasswordURL = "/password-confirm"
	confirmEmailURL    = "/email-confirm"
)

type jsonError struct {
	Error string `json:"error"`
}

func newJsonError(err error) jsonError {
	return jsonError{err.Error()}
}

func sendConfirmationEmail(
	to, passwordhash string,
	urlpath string,
	resetPassword bool) error {
	cfg := config.Get()
	token, err := apptoken.NewEmailTokenString(
		cfg.OAuth2State.TokenIssuer,
		to,
		passwordhash,
		cfg.ConfirmationLinkLifespan,
		cfg.OAuth2State.TokenSignKey)
	if err != nil {
		return err
	}

	externalURL := cfg.ExternalBaseURL + urlpath
	link := fmt.Sprintf("%s?token=%s", externalURL, token)
	var text, topic string
	if resetPassword {
		text = fmt.Sprintf("Follow this link to change your password: %s\n", link)
		topic = "Confirm password reset"
	} else {
		text = fmt.Sprintf("Follow this link to confirm your email address and complete creating account: %s\n", link)
		topic = "Confirm account creation"
	}
	return email.Send(to, topic, text)
}
