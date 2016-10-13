package handler

import (
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/config"
)

// NewHandler returns default Handler implemetation.
func NewHandler(
	c config.Config,
	ec ErrorCustomizer,
	as AuthService,
	us UserService) Handler {
	return handler{c, ec, as, us}
}

//TODO: describe API in swagger format.

// Handler combines http-handlers useful to create login logic.
type Handler interface {

	// AuthCodeURLs responds with auth code URLs for OAuth2.
	// Handler takes slice of oauth2.Config from configuration supplied
	// to NewHandler() func and renders list of URLs to response body.
	// Web UI could use this request to update its list of providers with fresh
	// URLs (re-generate state query parameter in them).
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
}

//TODO: currently handler marshals response as JSON; we may provide setting
// (marshalling func in config) to change response type (for ex. c.XML()).

// handler implements Handler interface.
// Note: methods are implemented in separate files.
type handler struct {
	config          config.Config
	errorCustomizer ErrorCustomizer
	auth            AuthService
	users           UserService
}
