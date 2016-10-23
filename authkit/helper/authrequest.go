package helper

import (
	"github.com/letsrock-today/hydra-sample/authkit"
	"golang.org/x/oauth2"
)

func WithOAuthTokenDo(us authkit.MiddlewareUserService, login, providerID string, do func(token *oauth2.Token) (*oauth2.Token, error)) error {
	token, err := us.OAuth2Token(login, providerID)
	if err != nil {
		return err
	}
	newToken, err := do(token)
	if err != nil {
		return err
	}
	if newToken != nil && newToken != token {
		return us.UpdateOAuth2Token(login, providerID, newToken)
	}
	return nil
}
