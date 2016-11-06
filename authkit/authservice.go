package authkit

import "golang.org/x/oauth2"

type (

	// AuthService provides a low-level auth implemetation.
	AuthService interface {
		HandlerAuthService
		TokenValidator
	}

	// HandlerAuthService provides a low-level auth implemetation,
	// specific to the handler package.
	// Interface is based on Hydra OAuth2 provider and may not fit to be used
	// with another providers, of which we are currently are not aware.
	HandlerAuthService interface {

		// GenerateConsentToken returns consent token used to create redirect
		// URL, which is used to redirect to OAuth2 backend from consent page.
		GenerateConsentToken(
			subj string,
			scopes []string,
			challenge string) (string, error)

		// GenerateConsentTokenPriv returns consent token used to create redirect
		// URL, which is used to redirect to OAuth2 backend from web app's own
		// login page (that is suffix "Priv" means "private" or "privileged" use).
		GenerateConsentTokenPriv(
			subj string,
			scopes []string,
			clientID string) (string, error)

		// IssueToken returns a token from OAuth2 backend for own web app
		// in case of form-based login within app.
		// Used for already authorized users.
		IssueToken(login string) (*oauth2.Token, error)

		// RevokeAccessToken revokes access token.
		RevokeAccessToken(accessToken string) error
	}

	// TokenValidator used to validate token and to check if token has
	// required permission. TokenValidator works in pair with PermissionMapper.
	// PermissionMapper takes http method and path and converts them to
	// permissionDescriptor, which TokenValidator used to validate token against.
	// Type of permissionDescriptor is specific to application's OAuth2 provider.
	TokenValidator interface {

		// Validate checks if token is valid and has required permission.
		// Valideate returns subject related to provided access token.
		Validate(
			accessToken string,
			permissionDescriptor interface{}) (subj string, err error)
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

//go:generate mockery -name AuthService
