package authkit

// ErrorCustomizer used to transform low-level errors before render them to
// http response. It could be used to provide app-specific error code or i18n.
// Values returned by interface methods are marshalled into response.
// Note, that with current implementation, JSON marshaller used to marshall
// errors. Errors may be any structures, not necessary implement error interface.
// ErrorCustomizer implementation should hide low-level details (stack trace,
// technical error details) from the end user.
type ErrorCustomizer interface {
	InvalidRequestParameterError(error) interface{}
	UserCreationError(error) interface{}
	UserAuthenticationError(error) interface{}
}
