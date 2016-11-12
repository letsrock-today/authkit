package profile

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/profile"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/socialprofile"
	"github.com/pkg/errors"
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
			"login": bson.M{
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
		Key:    []string{"login"},
		Unique: true,
	}
	return s, s.profiles.EnsureIndex(index)
}

func (s service) Profile(login string) (authkit.Profile, error) {
	p := socialprofile.Profile{}
	err := s.profiles.Find(
		bson.M{
			"login": login,
		}).One(&p)
	return &p, err
}

func (s service) Save(p authkit.Profile) error {
	_, err := s.profiles.Upsert(
		bson.M{
			"login": p.GetLogin(),
		},
		p)
	return err
}

func (s service) EnsureExists(login, email string) error {
	_, err := s.profiles.Upsert(
		bson.M{
			"login": login,
		},
		bson.M{
			"$setOnInsert": bson.M{
				"login": login,
				"email": email,
			},
		})
	if err == mgo.ErrNotFound {
		return errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	return err
}

func (s service) SetEmailConfirmed(login, email string, confirmed bool) error {
	err := s.profiles.Update(
		bson.M{
			"login": login,
			"email": email,
		},
		bson.M{
			"$set": bson.M{
				"emailconfirmed": confirmed,
			},
		})
	if err == mgo.ErrNotFound {
		return errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	return err
}

func (s service) Email(login string) (string, string, error) {
	p := socialprofile.Profile{}
	err := s.profiles.Find(
		bson.M{
			"login": login,
		}).One(&p)
	if err == mgo.ErrNotFound {
		return "", "", errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	if err != nil {
		return "", "", err
	}
	return p.Email, p.FormattedName, nil
}

func (s service) ConfirmedEmail(login string) (string, string, error) {
	p := socialprofile.Profile{}
	err := s.profiles.Find(
		bson.M{
			"login":          login,
			"emailconfirmed": true,
		}).One(&p)
	if err == mgo.ErrNotFound {
		return "", "", errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	if err != nil {
		return "", "", err
	}
	return p.Email, p.FormattedName, nil
}

func (s service) Close() error {
	s.dbsession.Close()
	return nil
}
