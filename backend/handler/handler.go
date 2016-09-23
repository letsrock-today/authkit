package handler

import (
	"fmt"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/service/socialprofile"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
	"github.com/letsrock-today/hydra-sample/backend/util/email"
)

var (
	Users    userapi.UserAPI
	Profiles socialprofile.ProfileAPI
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
	token, err := newStateToken(
		cfg.OAuth2State.TokenSignKey,
		cfg.OAuth2State.TokenIssuer,
		to,
		passwordhash,
		cfg.ConfirmationLinkLifespan)
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
