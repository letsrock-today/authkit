package middleware

import (
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/labstack/echo"
)

type (

	// TokenValidator used to validate token and to check if token has required permission.
	TokenValidator interface {

		// Validate checks if token is valid and has required permission.
		Validate(accessToken string, permissionDescriptor interface{}) error
	}

	// UserStore allows to find user data by access token.
	// If access token passed validation, user data stores in the echo.Context.
	// Also, UserStore may update refreshed OAuth2 token.
	UserStore interface {

		// User returns user data by access token.
		User(accessToken string) (interface{}, error)

		// OAuth2Token retrieves oauth2 token for user from store (or from user data).
		OAuth2Token(user interface{}) (*oauth2.Token, error)

		// UpdateOAuth2Token saves oauth2 token in the store.
		UpdateOAuth2Token(user interface{}, token *oauth2.Token) error

		// UserContext returns user data to be stored in the echo.Context.
		// It may return same structure which is passed to it or some fields from it.
		UserContext(user interface{}) interface{}
	}

	// TokenSourceProvider is implemented by oauth2.Config.
	// This interface extracted for testability.
	TokenSourceProvider interface {
		TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource
	}

	// AccessTokenConfig is a configuration for AccessTokenWithConfig middleware.
	AccessTokenConfig struct {

		// Context key to store user info into context.
		// Optional. Default value is "user-context".
		ContextKey string

		// Optional. Default value is DefaultPermissionMapper
		PermissionMapper PermissionMapper

		// Required.
		TokenValidator TokenValidator

		// Required.
		UserStore UserStore

		// Config used to refresh OAuth2 token.
		// Optional. Default value is nil, which disables token refresh.
		OAuth2Config TokenSourceProvider

		// Context used to refresh OAuth2 token.
		// Optional. Default value is nil, which disables token refresh.
		OAuth2Context context.Context
	}
)

var (
	// for use by tests
	reportEffectiveConfig func(AccessTokenConfig) = nil
)

const (
	invalidHeaderFormatMsg = "invalid header format"
	notPermittedMsg        = "operation is not permitted"

	DefaultContextKey = "user-context"
)

func NewDefaultPermissionMapper() PermissionMapper {
	return DefaultPermissionMapper{}
}

func AccessToken(us UserStore, tv TokenValidator) echo.MiddlewareFunc {
	c := AccessTokenConfig{
		ContextKey:       DefaultContextKey,
		PermissionMapper: NewDefaultPermissionMapper(),
	}
	c.UserStore = us
	c.TokenValidator = tv
	return AccessTokenWithConfig(c)
}

func AccessTokenWithConfig(config AccessTokenConfig) echo.MiddlewareFunc {
	// Defaults
	if config.ContextKey == "" {
		config.ContextKey = DefaultContextKey
	}
	if config.PermissionMapper == nil {
		config.PermissionMapper = NewDefaultPermissionMapper()
	}
	// Required
	if config.UserStore == nil {
		panic("UserStore must be provided")
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
				return echo.NewHTTPError(http.StatusForbidden, invalidHeaderFormatMsg)
			}
			token := strings.TrimSpace(split[1])
			if token == "" {
				return echo.NewHTTPError(http.StatusForbidden, invalidHeaderFormatMsg)
			}

			// Map request to permission
			perm, err := config.PermissionMapper.RequiredPermissioin(req.Method(), req.URL().Path())
			if err != nil {
				log.Println(err)
				return echo.NewHTTPError(http.StatusForbidden, notPermittedMsg)
			}

			// Find user
			user, err := config.UserStore.User(token)
			if err != nil {
				log.Println(err)
				return echo.NewHTTPError(http.StatusForbidden, notPermittedMsg)
			}

			// Update OAuth2 token and save it in DB
			err = func() error {
				if config.OAuth2Config != nil && config.OAuth2Context != nil {
					oauth2token, err := config.UserStore.OAuth2Token(user)
					if err != nil {
						return err
					}
					newToken, err := config.OAuth2Config.TokenSource(config.OAuth2Context, oauth2token).Token()
					if err != nil {
						return err
					}
					if newToken != oauth2token {
						err = config.UserStore.UpdateOAuth2Token(user, newToken)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}()
			if err != nil {
				log.Println(err)
			}

			// Validate token's permissions
			if err := config.TokenValidator.Validate(token, perm); err != nil {
				log.Println(err)
				return echo.NewHTTPError(http.StatusForbidden, notPermittedMsg)
			}

			// Store user login to context.
			c.Set(config.ContextKey, config.UserStore.UserContext(user))
			return next(c)
		}
	}
}
