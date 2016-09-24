package profile

import (
	"log"

	api "github.com/letsrock-today/hydra-sample/backend/service/profile/profileapi"
	"github.com/letsrock-today/hydra-sample/backend/service/socialprofile"
)

type profileapi struct{}

func New() api.ProfileAPI {
	return profileapi{}
}

func (profileapi) Profile(login string) (*socialprofile.Profile, error) {
	//TODO: use DB
	return nil, nil
}

func (profileapi) Save(login string, profile *socialprofile.Profile) error {
	//TODO: use DB
	log.Println(profile)
	return nil
}

func (profileapi) Close() error {
	//hs.dbsession.Close()
	return nil
}
