package handler

import (
	"golang.org/x/oauth2"

	"github.com/letsrock-today/authkit/authkit"
)

// userTokenStore retrieves tokens from user, saving one database request.
// If supplied user doesn't provide method OAuth2TokenByProviderID(),
// then userTokenStore fall back to authkit.TokenStore implementation, which is
// to retrieve token from store with separate request.
type userTokenStore struct {
	authkit.TokenStore
	u authkit.User
}

func (uts userTokenStore) OAuth2Token(
	login, providerID string) (*oauth2.Token, authkit.UserServiceError) {
	if login != uts.u.Login() {
		panic("illegal state")
	}
	type user interface {
		OAuth2TokenByProviderID(string) *oauth2.Token
	}
	if u, ok := uts.u.(user); ok {
		return u.OAuth2TokenByProviderID(providerID), nil
	}
	return uts.TokenStore.OAuth2Token(login, providerID)
}
