package profile

import (
	"io"

	"github.com/letsrock-today/authkit/authkit"
)

// Service represents locally stored user's profile.
type Service interface {
	io.Closer
	authkit.ProfileService
	Profile(login string) (authkit.Profile, error)
}
