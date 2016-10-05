package profileapi

import (
	"io"

	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/socialprofile"
)

type ProfileAPI interface {

	// close storage
	io.Closer

	// context should be created by oauth2 and contain token
	Profile(login string) (*socialprofile.Profile, error)

	Save(login string, profile *socialprofile.Profile) error
}
