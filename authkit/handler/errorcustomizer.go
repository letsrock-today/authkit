package handler

// ErrorCustomizer used to transform low-level errors before render them to
// http response. It could be used to provide app-specific error code or i18n.
// Values returned by interface methods are marshalled into response.
type ErrorCustomizer interface {
	InvalidRequestParameterError(error) interface{}
	UserCreationError(error) interface{}
	UserAuthenticationError(error) interface{}
}
