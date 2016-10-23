package confirmer

import (
	"fmt"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
)

type confirmer struct {
	confirmEmailURL    string
	confirmPasswordURL string
}

func New(confirmEmailURL, confirmPasswordURL string) authkit.Confirmer {
	return confirmer{
		confirmEmailURL,
		confirmPasswordURL,
	}
}

func (c confirmer) RequestEmailConfirmation(
	login string) authkit.UserServiceError {
	err := sendConfirmationEmail(login, "", c.confirmEmailURL, false)
	if err != nil {
		return authkit.NewRequestConfirmationError(err)
	}
	return nil
}

func (c confirmer) RequestPasswordChangeConfirmation(
	login, passwordHash string) authkit.UserServiceError {
	err := sendConfirmationEmail(login, passwordHash, c.confirmPasswordURL, true)
	if err != nil {
		return authkit.NewRequestConfirmationError(err)
	}
	return nil
}

func sendConfirmationEmail(
	to, passwordhash string,
	urlpath string,
	resetPassword bool) error {
	c := config.Get()
	oauth2State := c.OAuth2State()
	token, err := apptoken.NewEmailTokenString(
		oauth2State.TokenIssuer(),
		to,
		passwordhash,
		c.ConfirmationLinkLifespan(),
		oauth2State.TokenSignKey())
	if err != nil {
		return err
	}

	externalURL := c.ExternalBaseURL() + urlpath
	link := fmt.Sprintf("%s?token=%s", externalURL, token)
	var text, topic string
	if resetPassword {
		text = fmt.Sprintf("Follow this link to change your password: %s\n", link)
		topic = "Confirm password reset"
	} else {
		text = fmt.Sprintf("Follow this link to confirm your email address and complete creating account: %s\n", link)
		topic = "Confirm account creation"
	}
	return Send(to, topic, text)
}
