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
	Login        string
	PasswordHash string
	Tokens       map[string]*oauth2.Token // pid -> token
}

type _user struct {
	// embedded structure used to avoid clash of field and method names
	// and still export fields for BSON marshalling
	data userData
}

func (u *_user) Login() string {
	return u.data.Login
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
			"login": bson.M{
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
		Key:    []string{"login"},
		Unique: true,
	}
	return s, s.users.EnsureIndex(index)
}

func (s store) Close() error {
	s.dbsession.Close()
	return nil
}

func (s store) Create(login, password string) authkit.UserServiceError {
	err := s.users.Insert(
		&_user{
			userData{
				Login:        login,
				PasswordHash: hash(password),
			},
		})
	if mgo.IsDup(err) {
		return errors.WithStack(authkit.NewDuplicateUserError(err))
	}
	if err == mgo.ErrNotFound {
		return errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	return err
}

func (s store) Authenticate(login, password string) authkit.UserServiceError {
	u := _user{}
	err := s.users.Find(
		bson.M{
			"login":        login,
			"passwordhash": hash(password),
		}).One(&u)
	if err == mgo.ErrNotFound {
		return errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	return err
}

func (s store) User(login string) (authkit.User, authkit.UserServiceError) {
	u := &_user{}
	err := s.users.Find(
		bson.M{
			"login": login,
		}).One(u)
	if err == mgo.ErrNotFound {
		return nil, errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	return u, nil
}

func (s store) UpdatePassword(
	login, oldPasswordHash, newPassword string) authkit.UserServiceError {
	err := s.users.Update(
		bson.M{
			"login":        login,
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

func (s store) OAuth2TokenAndLoginByAccessToken(
	accessToken, providerID string) (*oauth2.Token, string, authkit.UserServiceError) {
	u := &_user{}
	err := s.users.Find(
		bson.M{
			"tokens." + providerID + ".accesstoken": accessToken,
		}).One(u)
	if err == mgo.ErrNotFound {
		return nil, "", errors.WithStack(authkit.NewUserNotFoundError(err))
	}
	return u.OAuth2TokenByProviderID(providerID), u.Login(), err
}

func (s store) UpdateOAuth2Token(
	login, providerID string, token *oauth2.Token) authkit.UserServiceError {
	err := s.users.Update(
		bson.M{
			"login": login,
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

func (s store) RevokeAccessToken(
	providerID, accessToken string) authkit.UserServiceError {
	err := s.users.Update(
		bson.M{
			"tokens." + providerID + ".accesstoken": accessToken,
		},
		bson.M{
			"$set": bson.M{
				"tokens." + providerID + ".accesstoken": nil,
			},
		})
	if err == mgo.ErrNotFound {
		return nil
	}
	return err
}

func (s store) Principal(u authkit.User) interface{} {
	return u
}

func hash(pass string) string {
	h := md5.New()
	h.Write([]byte(pass))
	return fmt.Sprintf("%x", h.Sum([]byte{}))
}
