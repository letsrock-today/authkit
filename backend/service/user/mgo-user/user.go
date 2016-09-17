package user

import (
	"crypto/md5"
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	api "github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

const (
	dbURL              = "localhost"
	dbName             = "hydra-sample"
	userCollectionName = "users"
)

type userapi struct {
	dbsession *mgo.Session
	users     *mgo.Collection
}

func New() (api.UserAPI, error) {
	s, err := mgo.Dial(dbURL)
	if err != nil {
		return nil, err
	}
	ua := &userapi{
		dbsession: s,
		users:     s.DB(dbName).C(userCollectionName),
	}
	err = ua.users.Create(&mgo.CollectionInfo{
		Validator: bson.M{
			"email":        bson.M{"$exists": true},
			"passwordhash": bson.M{"$exists": true},
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
	return ua, ua.users.EnsureIndex(index)
}

func (ua userapi) Close() error {
	ua.dbsession.Close()
	return nil
}

func (ua userapi) Create(login, password string) error {
	err := ua.users.Insert(
		&api.User{
			Email:        login,
			PasswordHash: hash(password),
		})
	if mgo.IsDup(err) {
		return api.AuthErrorDup
	}
	return err
}

func (ua userapi) Authenticate(login, password string) error {
	user := api.User{}
	err := ua.users.Find(
		bson.M{
			"email":        login,
			"passwordhash": hash(password),
		}).One(&user)
	if err != nil {
		if err == mgo.ErrNotFound {
			return api.AuthError
		}
		return err
	}
	return nil
}

func (ua userapi) GetUser(email string) (api.User, error) {
	// TODO
	return api.User{}, nil
}

func hash(pass string) string {
	h := md5.New()
	h.Write([]byte(pass))
	return fmt.Sprintf("%x", h.Sum([]byte{}))
}
