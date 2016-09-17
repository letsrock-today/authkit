package userapi

import (
	"errors"
	"io"
)

type UserAPI interface {
	io.Closer
	Create(login, password string) error
	Authenticate(login, password string) error
	GetUser(email string) (User, error)
}

type User struct {
	Email        string
	PasswordHash string
}

var (
	AuthError    = errors.New("Authentication error (invalid credentials?)")
	AuthErrorDup = errors.New("Authentication error (user already exists)")
)
