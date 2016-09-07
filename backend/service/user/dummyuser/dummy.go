package dummyuser

import (
	"errors"
	api "github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

type dummyuserapi struct{}

func New() api.UserAPI {
	return dummyuserapi{}
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
