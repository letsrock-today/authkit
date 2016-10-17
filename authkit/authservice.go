package authkit

import (
	"context"

	"golang.org/x/oauth2"
)

type (

	// AuthService provides a low-level auth implemetation.
	AuthService interface {
		HandlerAuthService
		PermissionMapper
	}

	// HandlerAuthService provides a low-level auth implemetation,
	// specific to the handler package.
	HandlerAuthService interface {
		GenerateConsentToken(
			subj string,
			scopes []string,
			challenge string) (string, error)
		IssueConsentToken(
			clientID string,
			scopes []string) (string, error)
		IssueToken(c context.Context, login string) (*oauth2.Token, error)
	}

	// TokenValidator used to validate token and to check if token has
	// required permission. TokenValidator works in pair with PermissionMapper.
	// PermissionMapper takes http method and path and converts them to
	// permissionDescriptor, which TokenValidator used to validate token against.
	// Type of permissionDescriptor is specific to application's OAuth2 provider.
	TokenValidator interface {

		// Validate checks if token is valid and has required permission.
		Validate(accessToken string, permissionDescriptor interface{}) error
	}

	// PermissionMapper used to map method and path of http request to desirable
	// permission descriptor. Permission descriptor is an interface, passed to
	// the TokenValidator. For example, in case of Hydra-backed TokenValidator,
	// permission descriptor contains resource name, action and scope.
	PermissionMapper interface {

		// RequiredPermissioin returns permission descriptor to be passed to
		// TokenValidator. It may return an error to prevent access to resource
		// without request to TokenValidator.
		RequiredPermissioin(method, path string) (interface{}, error)
	}
)
