package userapi

import (
	"errors"
	"io"
	"time"

	"golang.org/x/oauth2"
)

type UserAPI interface {

	// close storage
	io.Closer

	// create new disabled user
	Create(login, password string) error

	// authenticate user, returns nil, if account exists and enabled
	Authenticate(login, password string) error

	// return user by login
	User(login string) (*User, error)

	// update password
	UpdatePassword(login, password string) error

	// enable account
	Enable(login string) error

	// save token
	UpdateToken(login, pid string, token *oauth2.Token) error

	// return token
	Token(login, pid string) (*oauth2.Token, error)

	// return user by token
	// token_field is one of [accesstoken, refreshtoken]
	UserByToken(pid, token_field, token string) (*User, error)
}

type User struct {
	Email        string
	PasswordHash string
	Disabled     *time.Time               `bson:",omitempty"`
	Tokens       map[string]*oauth2.Token // pid -> token
}

var (
	AuthError             = errors.New("Authentication error (invalid credentials?)")
	AuthErrorDup          = errors.New("Authentication error (user already exists)")
	AuthErrorUserNotFound = errors.New("Authentication error (user not found)")
	AuthErrorDisabled     = errors.New("Authentication error (account disabled)")
)
