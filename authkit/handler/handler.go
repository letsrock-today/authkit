package handler

import (
	"github.com/asaskevich/govalidator"
	"github.com/letsrock-today/authkit/authkit"
)

// Config holds configuration parmeters for handler.NewHandler().
type Config struct {
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
func NewHandler(
	ac authkit.Config,
	hc Config) authkit.Handler {
	if !hc.Valid() {
		panic("invalid argument")
	}
	if hc.ContextCreator == nil {
		hc.ContextCreator = authkit.DefaultContextCreator{}
	}
	if hc.PasswordValidator == nil {
		hc.PasswordValidator = govalidator.Validator(defaultPasswordValidator)
	}
	govalidator.TagMap["password"] = hc.PasswordValidator
	if hc.LoginValidator == nil {
		hc.LoginValidator = govalidator.Validator(emailOrLoginValidator)
	}
	govalidator.TagMap["login"] = hc.LoginValidator
	return handler{
		ac,
		hc.ErrorCustomizer,
		hc.AuthService,
		hc.UserService,
		hc.ProfileService,
		hc.SocialProfileServices,
		hc.ContextCreator,
	}
}

// handler implements Handler interface.
// Note: methods are implemented in separate files.
type handler struct {
	config          authkit.Config
	errorCustomizer authkit.ErrorCustomizer
	auth            authkit.HandlerAuthService
	users           authkit.HandlerUserService
	profiles        authkit.ProfileService
	socialProfiles  authkit.SocialProfileServices
	contextCreator  authkit.ContextCreator
}
