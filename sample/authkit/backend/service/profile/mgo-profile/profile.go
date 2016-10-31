package profile

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/profile"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/socialprofile"
)

type service struct {
	dbsession *mgo.Session
	profiles  *mgo.Collection
}

// New creates new profile.Service based on MongoDB.
func New(dbURL, dbName, profileCollectionName string) (profile.Service, error) {
	ss, err := mgo.Dial(dbURL)
	if err != nil {
		return nil, err
	}
	s := &service{
		dbsession: ss,
		profiles:  ss.DB(dbName).C(profileCollectionName),
	}
	err = s.profiles.Create(&mgo.CollectionInfo{
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
	return s, s.profiles.EnsureIndex(index)
}

func (s service) Profile(login string) (authkit.Profile, error) {
	p := socialprofile.Profile{}
	err := s.profiles.Find(
		bson.M{
			"email": login,
		}).One(&p)
	return &p, err
}

func (s service) Save(p authkit.Profile) error {
	_, err := s.profiles.Upsert(
		bson.M{
			"email": p.Login(),
		},
		p)
	return err
}

func (s service) EnsureExists(login string) error {
	_, err := s.profiles.Upsert(
		bson.M{
			"email": login,
		},
		bson.M{
			"$setOnInsert": bson.M{
				"email": login,
			},
		})
	return err
}

func (s service) Close() error {
	s.dbsession.Close()
	return nil
}
