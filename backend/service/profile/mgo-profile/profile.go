package profile

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	api "github.com/letsrock-today/hydra-sample/backend/service/profile/profileapi"
	"github.com/letsrock-today/hydra-sample/backend/service/socialprofile"
)

//TODO: pass as a parameters to New()
const (
	dbURL                 = "localhost"
	dbName                = "hydra-sample"
	profileCollectionName = "profiles"
)

type profileapi struct {
	dbsession *mgo.Session
	profiles  *mgo.Collection
}

func New() (api.ProfileAPI, error) {
	s, err := mgo.Dial(dbURL)
	if err != nil {
		return nil, err
	}
	pa := &profileapi{
		dbsession: s,
		profiles:  s.DB(dbName).C(profileCollectionName),
	}
	err = pa.profiles.Create(&mgo.CollectionInfo{
		Validator: bson.M{
			"email": bson.M{
				"$exists": true,
				"$ne":     "",
			},
		},
	})
	// unfortunately, there is no other field to distinguish this error
	if err != nil && err.Error() != "collection already exists" {
		return nil, err
	}
	index := mgo.Index{
		Key:    []string{"email"},
		Unique: true,
	}
	return pa, pa.profiles.EnsureIndex(index)
}

func (pa profileapi) Profile(login string) (*socialprofile.Profile, error) {
	profile := socialprofile.Profile{}
	err := pa.profiles.Find(
		bson.M{
			"email": login,
		}).One(&profile)
	return &profile, err
}

func (pa profileapi) Save(login string, profile *socialprofile.Profile) error {
	_, err := pa.profiles.Upsert(
		bson.M{
			"email": login,
		},
		profile)
	return err
}

func (pa profileapi) Close() error {
	pa.dbsession.Close()
	return nil
}
