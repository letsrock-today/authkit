package handler

import (
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/profile/profileapi"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user/userapi"
)

var (
	Users    userapi.UserAPI
	Profiles profileapi.ProfileAPI
)

type jsonError struct {
	Error string `json:"error"`
}

func newJsonError(err error) jsonError {
	return jsonError{err.Error()}
}
