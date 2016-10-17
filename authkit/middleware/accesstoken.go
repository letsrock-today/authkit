package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/pkg/errors"
)

type (

	// AccessTokenConfig is a configuration for AccessTokenWithConfig middleware.
	AccessTokenConfig struct {

		// OAuth2 provider ID used to store and retrieve token from UserService.
		// Required.
		PrivateProviderID string

		// Context key to store user info into context.
		// Optional. Default value is "user-context".
		ContextKey string

		// Optional. Default value is DefaultPermissionMapper
		PermissionMapper authkit.PermissionMapper

		// Required.
		TokenValidator authkit.TokenValidator

		// Required.
		UserService authkit.MiddlewareUserService

		// Config used to refresh OAuth2 token.
		// Optional. Default value is nil, which disables token refresh.
		OAuth2Config authkit.TokenSourceProvider

		// ContextCreator used to obtain context to store and refresh OAuth2 token.
		// Optional. Default value is nil, which disables token refresh.
		ContextCreator authkit.ContextCreator
	}
)

var (
	// for use by tests
	errInvalidAuthHeader  = echo.NewHTTPError(http.StatusForbidden, "invalid header format")
	errAccessDenied       = echo.NewHTTPError(http.StatusForbidden, "access denied")
	reportEffectiveConfig func(AccessTokenConfig)
)

// DefaultContextKey declares default key which is used to store principal in the echo.Context.
const DefaultContextKey = "user-context"

// AccessToken used to create middleware with mostly default configuration.
func AccessToken(
	privateProviderID string,
	us authkit.MiddlewareUserService,
	tv authkit.TokenValidator) echo.MiddlewareFunc {
	c := AccessTokenConfig{
		PrivateProviderID: privateProviderID,
		ContextKey:        DefaultContextKey,
		PermissionMapper:  DefaultPermissionMapper{},
	}
	c.UserService = us
	c.TokenValidator = tv
	return AccessTokenWithConfig(c)
}

// AccessTokenWithConfig used to create middleware with provided configuration.
func AccessTokenWithConfig(config AccessTokenConfig) echo.MiddlewareFunc {
	// Defaults.
	if config.ContextKey == "" {
		config.ContextKey = DefaultContextKey
	}
	if config.PermissionMapper == nil {
		config.PermissionMapper = DefaultPermissionMapper{}
	}
	// Required.
	if config.UserService == nil {
		panic("UserService must be provided")
	}
	if config.TokenValidator == nil {
		panic("TokenValidator must be provided")
	}
	if reportEffectiveConfig != nil {
		reportEffectiveConfig(config)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Get access token from header.

			auth := req.Header().Get("Authorization")
			split := strings.SplitN(auth, " ", 2)
			if len(split) != 2 || !strings.EqualFold(split[0], "bearer") {
				return errInvalidAuthHeader
			}
			token := strings.TrimSpace(split[1])
			if token == "" {
				return errInvalidAuthHeader
			}

			// Map request to permission.
			perm, err := config.PermissionMapper.RequiredPermissioin(
				req.Method(),
				req.URL().Path())
			if err != nil {
				c.Logger().Debug(errors.WithStack(err))
				return errAccessDenied
			}

			// Find user.
			user, err := config.UserService.UserByAccessToken(token)
			if err != nil {
				c.Logger().Debug(errors.WithStack(err))
				return errAccessDenied
			}

			// Update OAuth2 token and save it in DB (asynchronously).
			// Possible errors are irrelevant for the main code flow.
			sync := make(chan struct{}, 1)
			if config.OAuth2Config != nil && config.ContextCreator != nil {
				go func() {
					defer func() { sync <- struct{}{} }()
					oauth2token, err := config.UserService.OAuth2Token(
						user.Login(),
						config.PrivateProviderID)
					if err != nil {
						c.Logger().Warn(errors.WithStack(err))
						return
					}
					newToken, err := config.OAuth2Config.TokenSource(
						config.ContextCreator.CreateContext(config.PrivateProviderID),
						oauth2token).Token()
					if err != nil {
						c.Logger().Warn(errors.WithStack(err))
						return
					}
					if newToken != oauth2token {
						err = config.UserService.UpdateOAuth2Token(
							user.Login(),
							config.PrivateProviderID,
							newToken)
						if err != nil {
							c.Logger().Warn(errors.WithStack(err))
						}
					}
				}()
			}

			// Validate token's permissions.
			if err := config.TokenValidator.Validate(token, perm); err != nil {
				c.Logger().Debug(errors.WithStack(err))
				return errAccessDenied
			}

			// Store user login to context.
			c.Set(config.ContextKey, config.UserService.Principal(user))

			// Make the whole call synchronouse.
			<-sync
			return next(c)
		}
	}
}
