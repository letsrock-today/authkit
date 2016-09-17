package user

import (
	"errors"
	api "github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

type dummyuserapi struct{}

func New() (api.UserAPI, error) {
	return &dummyuserapi{}, nil
}

func (dummyuserapi) Create(login, password string) error {
	return errors.New("Not implemented yet")
}

func (dummyuserapi) Authenticate(login, password string) error {
	if login != "zzz@zzz.zz" || password != "zzz" {
		err := errors.New("Invalid username and password combination")
		return err
	}
	return nil
}

func (dummyuserapi) GetUser(email string) (api.User, error) {
	return api.User{}, nil
}

func (dummyuserapi) Close() error {
	return nil
}
