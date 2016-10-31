package handler

import (
	"github.com/asaskevich/govalidator"
	"github.com/letsrock-today/authkit/authkit"
)

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
	errs, ok := e.(govalidator.Errors)
	if !ok {
		return []jsonError{{"invalid_req_param", e.Error()}}
	}
	ret := []jsonError{}
	for _, e := range errs {
		var je jsonError
		code := e.Error()
		switch code {
		// List of all possible custom validation errors.
		// This errors originated from tags on validated structures.
		case "login-required":
			je = jsonError{code, "Login is required"}
		case "login-format":
			je = jsonError{code, "Login should be a valid email address"}
		case "password-required":
			je = jsonError{code, "Password is required"}
		case "password-format":
			je = jsonError{code, `Password must be 5-50 chars long,
contain latin letters in both registers,
digits and other symbols (at least one of each kind)`}
		case "email-required":
			je = jsonError{code, "Email is required"}
		case "email-format":
			je = jsonError{code, "Email should be a valid email address"}
		default:
			je = jsonError{"invalid_req_param", code}
		}
		ret = append(ret, je)
	}
	return ret
}

func (ec) UserCreationError(e error) interface{} {
	msg := e.Error()
	switch {
	case authkit.IsAccountDisabled(e):
		return []jsonError{{"account_disabled", msg}}
	case authkit.IsDuplicateUser(e):
		return []jsonError{{"duplicate_account", msg}}
	}
	return []jsonError{{"unknown_err", msg}}
}

func (ec) UserAuthenticationError(e error) interface{} {
	return []jsonError{{"auth_err", e.Error()}}
}
