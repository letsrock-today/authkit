package middleware

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/persisttoken"
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

		// OAuth2Config used to refresh OAuth2 token.
		// Optional. Default value is nil, which disables token refresh.
		OAuth2Config authkit.OAuth2Config

		// ContextCreator used to obtain context to store and refresh OAuth2 token.
		// Optional. Default value is nil, which disables token refresh.
		ContextCreator authkit.ContextCreator

		// AuthHeaderName is a name of header to be used to update auth token on
		// the client.
		AuthHeaderName string

		// RefreshAllowedInterval is an interval since access token expiration
		// during which it is allowed to refresh expired access token.
		// It is not quite similar to traditional HTTP session timeout, but
		// used for the same purpose.
		RefreshAllowedInterval time.Duration
	}
)

var (
	// for use by tests
	errInvalidAuthHeader  = echo.NewHTTPError(http.StatusForbidden, "invalid header format")
	errAccessDenied       = echo.NewHTTPError(http.StatusForbidden, "access denied")
	reportEffectiveConfig func(AccessTokenConfig)
)

const (
	// DefaultContextKey declares default key which is used to store principal
	// in the echo.Context.
	DefaultContextKey = "user-context"

	// DefaultAuthHeaderName declares default header name used to update
	// access token on client web app.
	DefaultAuthHeaderName = "X-App-Auth"

	defaultRefreshAllowedInterval = 10 * time.Minute
)

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
	c.AuthHeaderName = DefaultAuthHeaderName
	c.RefreshAllowedInterval = defaultRefreshAllowedInterval
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
	if config.AuthHeaderName == "" {
		config.AuthHeaderName = DefaultAuthHeaderName
	}
	if config.RefreshAllowedInterval == 0 {
		config.RefreshAllowedInterval = defaultRefreshAllowedInterval
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
				c.Logger().Debugf("%+v", errors.WithStack(err))
				return errAccessDenied
			}

			// Validate token's permissions and retrieve subject (login).
			login, err := config.TokenValidator.Validate(token, perm)
			if err != nil {
				c.Logger().Debugf("%+v", errors.WithStack(err))
				if config.OAuth2Config == nil || config.ContextCreator == nil {
					c.Logger().Debugf("%+v", errors.New(
						"access token is invalid and refreshing is not enabled"))
					return errAccessDenied
				}
				// Update OAuth2 token and save it in DB.
				ctx := config.ContextCreator.CreateContext(
					config.PrivateProviderID)
				t, err1 := persisttoken.WrapOAuth2ConfigUseAccessToken(
					config.OAuth2Config,
					token,
					config.PrivateProviderID,
					config.UserService,
					func(t *oauth2.Token) error {
						// Restrict period of time during which we allow to
						// refresh token. This is not quite similar to
						// traditional HTTP session timeout, but for the same
						// purpose. If there were no requests to API since access
						// token expired and it expired more than this interval
						// ago, then we assume session inactive and not keep it.
						if !t.Valid() &&
							!t.Expiry.IsZero() &&
							time.Since(t.Expiry) > config.RefreshAllowedInterval {
							return errors.New("token expired too long ago")
						}
						return nil
					}).TokenSource(ctx, nil).Token()
				if err1 != nil {
					c.Logger().Debugf("%+v", errors.WithStack(err1))
					return errAccessDenied
				}
				if t.AccessToken == token {
					c.Logger().Debugf("%+v", errors.WithStack(err))
					return errAccessDenied
				}
				token = t.AccessToken
				login, err = config.TokenValidator.Validate(token, perm)
				if err != nil {
					c.Logger().Debugf("%+v", errors.WithStack(err))
					return errAccessDenied
				}
				// Access token is updated and valid.
				// Set custom header to update token on the client.
				c.Response().Header().Set(config.AuthHeaderName, token)
			}

			// Find user.
			user, err := config.UserService.User(login)
			if err != nil {
				c.Logger().Debugf("%+v", errors.WithStack(err))
				return errAccessDenied
			}

			// Map user to principal and put it into context.
			c.Set(config.ContextKey, config.UserService.Principal(user))

			return next(c)
		}
	}
}
