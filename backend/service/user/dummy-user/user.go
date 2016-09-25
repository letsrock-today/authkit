package user

import (
	"time"

	api "github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

type dummyuserapi struct {
	users map[string]api.User
}

func New() (api.UserAPI, error) {
	return &dummyuserapi{
		map[string]api.User{},
	}, nil
}

func (du *dummyuserapi) Close() error {
	du.users = map[string]api.User{}
	return nil
}

func (du *dummyuserapi) Create(login, password string) error {
	_, ok := du.users[login]
	if ok {
		return api.AuthErrorDup
	}
	t := time.Now()
	du.users[login] = api.User{
		Email:        login,
		PasswordHash: password,
		Disabled:     &t,
	}
	return api.AuthErrorDisabled
}

func (du *dummyuserapi) Authenticate(login, password string) error {
	u, ok := du.users[login]
	if !ok {
		return api.AuthError
	}
	if password != u.PasswordHash {
		return api.AuthError
	}
	if u.Disabled != nil {
		return api.AuthErrorDisabled
	}
	return nil
}

func (du *dummyuserapi) User(login string) (*api.User, error) {
	u, ok := du.users[login]
	if !ok {
		return nil, api.AuthErrorUserNotFound
	}
	return &u, nil
}

func (du *dummyuserapi) UpdatePassword(login, password string) error {
	u, err := du.User(login)
	if err != nil {
		return err
	}
	u.PasswordHash = password
	return nil
}

func (du *dummyuserapi) Enable(login string) error {
	u, err := du.User(login)
	if err != nil {
		return err
	}
	u.Disabled = nil
	return nil
}

func (du *dummyuserapi) UpdateToken(login, pid, token string) error {
	user, err := du.User(login)
	if err != nil {
		return err
	}
	user.Tokens[pid] = token
	return nil
}

func (du *dummyuserapi) Token(login, pid string) (string, error) {
	user, err := du.User(login)
	if err != nil {
		return "", err
	}
	return user.Tokens[pid], nil
}
