package user

import (
	"crypto/md5"
	"fmt"
	"time"

	"golang.org/x/oauth2"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/user"
)

type userData struct {
	Email        string
	PasswordHash string
	Disabled     *time.Time               `bson:",omitempty"`
	Tokens       map[string]*oauth2.Token // pid -> token
}

type _user struct {
	// embedded structure used to avoid clash of field and method names
	// and still export fields for BSON marshalling
	data userData
}

func (u *_user) Login() string {
	//TODO: remove assamption email == login?
	return u.data.Email
}

func (u *_user) Email() string {
	return u.data.Email
}

func (u *_user) PasswordHash() string {
	return u.data.PasswordHash
}

func (u *_user) OAuth2TokenByProviderID(providerID string) *oauth2.Token {
	return u.data.Tokens[providerID]
}

func (u *_user) GetBSON() (interface{}, error) {
	return u.data, nil
}

func (u *_user) SetBSON(raw bson.Raw) error {
	return raw.Unmarshal(&u.data)
}

type store struct {
	dbsession *mgo.Session
	users     *mgo.Collection
}

// New returns new user.Store based on MongoDB.
func New(
	dbURL, dbName, userCollectionName string,
	unconfirmedUserLifespan time.Duration) (user.Store, error) {
	ss, err := mgo.Dial(dbURL)
	if err != nil {
		return nil, err
	}
	s := &store{
		dbsession: ss,
		users:     ss.DB(dbName).C(userCollectionName),
	}
	err = s.users.Create(&mgo.CollectionInfo{
		Validator: bson.M{
			"email": bson.M{
				"$exists": true,
				"$ne":     "",
			},
			"passwordhash": bson.M{
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
	if err := s.users.EnsureIndex(index); err != nil {
		return nil, err
	}
	index = mgo.Index{
		Key:         []string{"disabled"},
		ExpireAfter: unconfirmedUserLifespan,
	}
	return s, s.users.EnsureIndex(index)
}

func (s store) Close() error {
	s.dbsession.Close()
	return nil
}

func (s store) Create(login, password string) authkit.UserServiceError {
	t := time.Now()
	err := s.users.Insert(
		&_user{
			userData{
				Email:        login,
				PasswordHash: hash(password),
				Disabled:     &t,
			},
		})
	if mgo.IsDup(err) {
		return errors.WithStack(authkit.NewDuplicateUserError(err))
	}
	if err == nil {
		return errors.WithStack(authkit.NewAccountDisabledError(nil))
	}
	return err
}

func (s store) CreateEnabled(login, password string) authkit.UserServiceError {
	err := s.users.Insert(
		&_user{
			userData{
				Email:        login,
				PasswordHash: hash(password),
				Disabled:     nil,
			},
		})
	if mgo.IsDup(err) {
		return errors.WithStack(authkit.NewDuplicateUserError(err))
	}
	return err
}

func (s store) Enable(login string) authkit.UserServiceError {
	err := s.users.Update(
		bson.M{
			"email": login,
		},
		bson.M{
			"$set": bson.M{
				"disabled": nil,
			},
		})
	if err == mgo.ErrNotFound {
		return errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	return err
}

func (s store) Authenticate(login, password string) authkit.UserServiceError {
	u := _user{}
	err := s.users.Find(
		bson.M{
			"email":        login,
			"passwordhash": hash(password),
		}).One(&u)
	if err != nil {
		if err == mgo.ErrNotFound {
			return errors.WithStack(authkit.NewUserNotFoundError(nil))
		}
		return err
	}
	if err == nil && u.data.Disabled != nil {
		return errors.WithStack(authkit.NewAccountDisabledError(nil))
	}
	return nil
}

func (s store) User(login string) (authkit.User, authkit.UserServiceError) {
	u := &_user{}
	err := s.users.Find(
		bson.M{
			"email": login,
		}).One(u)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.WithStack(authkit.NewUserNotFoundError(err))
		}
		return nil, err
	}
	return u, nil
}

func (s store) UpdatePassword(
	login, oldPasswordHash, newPassword string) authkit.UserServiceError {
	err := s.users.Update(
		bson.M{
			"email":        login,
			"passwordhash": oldPasswordHash,
		},
		bson.M{
			"$set": bson.M{
				"passwordhash": hash(newPassword),
			},
		})
	if err == mgo.ErrNotFound {
		return errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	return err
}

func (s store) OAuth2Token(
	login, providerID string) (*oauth2.Token, authkit.UserServiceError) {
	u, err := s.User(login)
	if err != nil {
		return nil, err
	}
	return u.(*_user).data.Tokens[providerID], nil
}

func (s store) UpdateOAuth2Token(
	login, providerID string, token *oauth2.Token) authkit.UserServiceError {
	err := s.users.Update(
		bson.M{
			"email": login,
		},
		bson.M{
			"$set": bson.M{
				"tokens." + providerID: token,
			},
		})
	if err == mgo.ErrNotFound {
		return errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	return err
}

func (s store) UserByAccessToken(
	providerID, accessToken string) (authkit.User, authkit.UserServiceError) {
	u := &_user{}
	err := s.users.Find(
		bson.M{
			"tokens." + providerID + ".accesstoken": accessToken,
		}).One(u)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.WithStack(authkit.NewUserNotFoundError(err))
		}
		return nil, err
	}
	if err == nil && u.data.Disabled != nil {
		return nil, errors.WithStack(authkit.NewAccountDisabledError(nil))
	}
	return u, nil

}

func (s store) Principal(u authkit.User) interface{} {
	return u
}

func hash(pass string) string {
	h := md5.New()
	h.Write([]byte(pass))
	return fmt.Sprintf("%x", h.Sum([]byte{}))
}
