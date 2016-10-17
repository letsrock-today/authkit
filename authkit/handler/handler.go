package handler

import (
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/config"
)

// NewHandler returns default Handler implemetation.
// ErrorCustomizer, AuthService, UserService, ProfileService must be provided.
// If ContextCreator is nil, then DefaultContextCreator is used.
func NewHandler(
	c config.Config,
	ec ErrorCustomizer,
	as AuthService,
	us UserService,
	ps ProfileService,
	sps SocialProfileServices,
	cc ContextCreator) Handler {
	if ec == nil || as == nil || us == nil || ps == nil || sps == nil {
		// Better to crash sooner.
		return nil
	}
	if cc == nil {
		cc = DefaultContextCreator{}
	}
	return handler{c, ec, as, us, ps, sps, cc}
}

type (
	//TODO: describe API in swagger format.

	// Handler combines http-handlers useful to create login logic.
	Handler interface {

		// AuthCodeURLs responds with auth code URLs for OAuth2.
		// Handler takes slice of oauth2.Config from configuration supplied
		// to NewHandler() func and renders list of URLs to response body.
		// Web UI could use this request to update its list of providers with
		// fresh URLs (re-generate state query parameter in them).
		// Response should not be cached.
		AuthCodeURLs(echo.Context) error

		// AuthProviders responds with list of OAuth2 providers, configured by the
		// application. Response could be used by web UI to represent a list of
		// providers with names and icons. Response could be cached.
		AuthProviders(echo.Context) error

		// ConsentLogin handles login requests from the consent page.
		ConsentLogin(echo.Context) error

		// Login handles login requests from the application's login page.
		Login(echo.Context) error

		// RestorePassword handles request to restore password
		// ("forgot password" link in the login form).
		RestorePassword(echo.Context) error

		// ChangePassword handles request to actually change password from the
		// confirmation form.
		ChangePassword(echo.Context) error

		// ConfirmEmail handles request to confirm email (which is produced
		// by the link sent to the user in the confirmation email).
		// Response for the user created with template named
		// "authkit.EmailConfirm.response". This template should be registered
		// in the echo.Context. I18n can be achieved with custom renderer.
		ConfirmEmail(echo.Context) error

		// Callback handles OAuth2 code flow callback requests.
		//Callback(echo.Context) error
	}

	//TODO: currently handler marshals response as JSON; we may provide setting
	// (marshalling func in config) to change response type (for ex. c.XML()).

	// handler implements Handler interface.
	// Note: methods are implemented in separate files.
	handler struct {
		config          config.Config
		errorCustomizer ErrorCustomizer
		auth            AuthService
		users           UserService
		profiles        ProfileService
		socialProfiles  SocialProfileServices
		contextCreator  ContextCreator
	}
)

// ConfirmEmailTemplateName is a template name for ConfirmEmail response.
const ConfirmEmailTemplateName = "authkit-ConfirmEmail-response.html"
