package handler

import (
	"github.com/asaskevich/govalidator"
	"github.com/letsrock-today/hydra-sample/authkit"
)

// NewHandler returns default Handler implemetation.
// All arguments except ContextCreator and Validator must be provided.
// If ContextCreator is nil, then DefaultContextCreator is used.
// If Validator is nil, then default password validator is used.
func NewHandler(
	c authkit.Config,
	ec authkit.ErrorCustomizer,
	as authkit.HandlerAuthService,
	us authkit.HandlerUserService,
	ps authkit.ProfileService,
	sps authkit.SocialProfileServices,
	cc authkit.ContextCreator,
	pv govalidator.Validator) authkit.Handler {
	if c == nil ||
		ec == nil ||
		as == nil ||
		us == nil ||
		ps == nil ||
		sps == nil {
		panic("invalid argument")
	}
	if cc == nil {
		cc = authkit.DefaultContextCreator{}
	}
	if pv == nil {
		pv = govalidator.Validator(defaultPasswordValidator)
	}
	govalidator.TagMap["password"] = pv
	return handler{c, ec, as, us, ps, sps, cc}
}

//TODO: currently handler marshals response as JSON; we may provide setting
// (marshalling func in config) to change response type (for ex. c.XML()).

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
