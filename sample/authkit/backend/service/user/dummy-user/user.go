package user

import (
	"errors"
	"time"

	"golang.org/x/oauth2"

	api "github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user/userapi"
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

func (du *dummyuserapi) UpdateToken(login, pid string, token *oauth2.Token) error {
	user, err := du.User(login)
	if err != nil {
		return err
	}
	user.Tokens[pid] = token
	return nil
}

func (du *dummyuserapi) Token(login, pid string) (*oauth2.Token, error) {
	user, err := du.User(login)
	if err != nil {
		return nil, err
	}
	return user.Tokens[pid], nil
}

func (du *dummyuserapi) UserByToken(pid, tokenField, token string) (*api.User, error) {
	for _, u := range du.users {
		t := u.Tokens[pid]
		switch tokenField {
		case "accesstoken":
			if t.AccessToken == token {
				return &u, nil
			}
		case "refreshtoken":
			if t.RefreshToken == token {
				return &u, nil
			}
		default:
			return nil, errors.New("unknown token type")
		}
	}
	return nil, api.AuthErrorUserNotFound
}
