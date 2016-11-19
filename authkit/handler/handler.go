package handler

import (
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/letsrock-today/authkit/authkit"
)

// Config holds configuration parmeters for handler.NewHandler().
type Config struct {

	// OAuth2Providers stores configuration of all registered external OAuth2
	// providers (except application's private OAuth2 provider).
	OAuth2Providers []authkit.OAuth2Provider

	// PrivateOAuth2Provider is a configuration of private OAuth2
	// provider. Private provider can be implemented by the app itself,
	// or it can be a third-party application. In both cases, it should
	// be available via http.
	PrivateOAuth2Provider authkit.OAuth2Provider

	// OAuth2State is a configuration of OAuth2 code flow state token.
	OAuth2State authkit.OAuth2State

	// AuthCookieName is a name of cookie to be used to send auth token to
	// the client.
	AuthCookieName string

	// ModTime is a configuration modification time. It is used to
	// cache list of providers on client (with "If-Modified_Since" header).
	ModTime time.Time

	ErrorCustomizer       authkit.ErrorCustomizer
	AuthService           authkit.HandlerAuthService
	UserService           authkit.HandlerUserService
	ProfileService        authkit.ProfileService
	SocialProfileServices authkit.SocialProfileServices
	ContextCreator        authkit.ContextCreator
	PasswordValidator     govalidator.Validator
	LoginValidator        govalidator.Validator
}

func (c Config) Valid() bool {
	return c.ErrorCustomizer != nil &&
		c.AuthService != nil &&
		c.UserService != nil &&
		c.ProfileService != nil &&
		c.SocialProfileServices != nil
}

// NewHandler returns default Handler implemetation.
// All arguments except ContextCreator and Validator must be provided.
// If ContextCreator is nil, then DefaultContextCreator is used.
// If Validator is nil, then default password validator is used.
func NewHandler(c Config) authkit.Handler {
	if !c.Valid() {
		panic("invalid argument")
	}
	if c.ContextCreator == nil {
		c.ContextCreator = authkit.DefaultContextCreator{}
	}
	if c.PasswordValidator == nil {
		c.PasswordValidator = govalidator.Validator(defaultPasswordValidator)
	}
	govalidator.TagMap["password"] = c.PasswordValidator
	if c.LoginValidator == nil {
		c.LoginValidator = govalidator.Validator(emailOrLoginValidator)
	}
	govalidator.TagMap["login"] = c.LoginValidator
	return handler{c}
}

// handler implements Handler interface.
// Note: methods are implemented in separate files.
type handler struct {
	Config
}
