package handler

import "github.com/letsrock-today/hydra-sample/backend/service/user/userapi"

var UserService userapi.UserAPI

type jsonError struct {
	Error string `json:"error"`
}

func newJsonError(err error) jsonError {
	return jsonError{err.Error()}
}
