package user

import (
	"io"

	"github.com/letsrock-today/hydra-sample/authkit"
)

// Store combines io.Closer and authkit.UserService.
type Store interface {
	io.Closer
	authkit.UserStore
}
