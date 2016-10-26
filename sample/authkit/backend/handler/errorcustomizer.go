package handler

import "github.com/letsrock-today/hydra-sample/authkit"

func NewErrorCustomizer() authkit.ErrorCustomizer {
	return ec{}
}

type jsonError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// This ErrorCustomizer implementation just illustrates basic idea.
// Idea is as follows:
// ErrorCustomizer groups all errors in several major groups and maps them to
// application-specific structure, which is marshalled to response.
// We use JSON, but may refactor authkit package to allow other types (XML, ...).
// Rich web-client app can use error codes to provide i18n error messages or
// fallback to unlocolized message.
type ec struct{}

func (ec) InvalidRequestParameterError(e error) interface{} {
	//TODO: use subcodes for concrete rules?
	return jsonError{"invalid_req_param", e.Error()}
}

func (ec) UserCreationError(e error) interface{} {
	switch {
	case authkit.IsAccountDisabled(e):
		return jsonError{"account_disabled", e.Error()}
	case authkit.IsDuplicateUser(e):
		return jsonError{"duplicate_account", e.Error()}
	}
	return jsonError{"unknown_err", e.Error()}
}

func (ec) UserAuthenticationError(e error) interface{} {
	return jsonError{"auth_err", e.Error()}
}
