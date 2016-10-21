package user

import (
	"io"

	"golang.org/x/oauth2"

	"github.com/letsrock-today/hydra-sample/authkit"
)

// User extends authkit.User with application-specific methods.
type User interface {
	authkit.User
	OAuth2TokenByProviderID(string) *oauth2.Token
}

// Store combines io.Closer and authkit.UserService.
type Store interface {
	io.Closer
	authkit.UserStore
}
