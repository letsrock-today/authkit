package helper

import (
	"golang.org/x/oauth2"

	"github.com/letsrock-today/hydra-sample/authkit"
)

// WithOAuth2TokenDo wraps usage of oauth2.Token. It retrieves token from the
// user storage before supplied callback execution and updates token in the
// storage after use. Callback should use provided token and return updated one.
func WithOAuth2TokenDo(
	ts authkit.TokenStore, login, providerID string,
	do func(token *oauth2.Token) (*oauth2.Token, error)) error {
	token, err := ts.OAuth2Token(login, providerID)
	if err != nil {
		return err
	}
	newToken, err := do(token)
	if err != nil {
		return err
	}
	if newToken != nil && newToken != token {
		return ts.UpdateOAuth2Token(login, providerID, newToken)
	}
	return nil
}
