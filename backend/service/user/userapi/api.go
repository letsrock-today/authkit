package userapi

import (
	"errors"
	"io"
	"time"
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
	UpdateToken(login, pid, token string) error

	// return token
	Token(login, pid string) (string, error)
}

type User struct {
	Email        string
	PasswordHash string
	Disabled     *time.Time        `bson:",omitempty"`
	Tokens       map[string]string // pid -> token
}

var (
	AuthError             = errors.New("Authentication error (invalid credentials?)")
	AuthErrorDup          = errors.New("Authentication error (user already exists)")
	AuthErrorUserNotFound = errors.New("Authentication error (user not found)")
	AuthErrorDisabled     = errors.New("Authentication error (account disabled)")
)
