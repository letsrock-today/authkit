package authkit

import "golang.org/x/oauth2"

type (

	// UserService provides methods to persist users.
	UserService interface {
		tokenStore
		middlewareMethods
		handlerMethods
	}

	// MiddlewareUserService provides methods to persist user.
	// Methods of this interface are specific to middleware package.
	MiddlewareUserService interface {
		tokenStore
		middlewareMethods
	}

	// HandlerUserService provides methods to persist user.
	// Methods of this interface are specific to handler package.
	HandlerUserService interface {
		tokenStore
		handlerMethods
	}

	tokenStore interface {
		// OAuth2Token returns OAuth2 token by login and OAuth2 provider ID.
		OAuth2Token(login, providerID string) (*oauth2.Token, UserServiceError)

		// UpdateOAuth2Token saves or updates oauth2 token for user and provider.
		UpdateOAuth2Token(login, providerID string, token *oauth2.Token) UserServiceError
	}

	middlewareMethods interface {
		// UserByAccessToken returns user data by access token.
		UserByAccessToken(accessToken string) (User, UserServiceError)

		// Principal returns user data to be stored in the echo.Context.
		// It may return same structure which is passed to it or some fields from it.
		Principal(user User) interface{}
	}

	handlerMethods interface {
		// Create creates new disabled user.
		Create(login, password string) UserServiceError

		// CreateEnabled creates new enabled user.
		CreateEnabled(login, password string) UserServiceError

		// Enable enables user account.
		Enable(login string) UserServiceError

		// Authenticate authenticates user, returns nil, if account exists and enabled.
		Authenticate(login, password string) UserServiceError

		// User returns user by login.
		User(login string) (User, UserServiceError)

		// UpdatePassword updates user's password.
		UpdatePassword(login, oldPasswordHash, newPassword string) UserServiceError

		// return user by token
		// token_field is one of [accesstoken, refreshtoken]
		//	UserByToken(pid, tokenField, token string) (User, UserServiceError)

		// RequestEmailConfirmation requests user to confirm email address.
		RequestEmailConfirmation(login string) UserServiceError

		// RequestPasswordChangeConfirmation requests user confirmation to change password (via email).
		RequestPasswordChangeConfirmation(login, passwordHash string) UserServiceError
	}

	// User provides basic information about user, required for login logic.
	User interface {
		Login() string
		Email() string
		PasswordHash() string
	}

	// UserServiceError is a general error specific to UserService.
	UserServiceError interface {
		error
	}

	// DuplicateUserError indicates that user already exists.
	DuplicateUserError interface {
		UserServiceError
		IsDuplicateUser() bool
	}

	// UserNotFoundError indicates that user not found.
	UserNotFoundError interface {
		UserServiceError
		IsUserNotFound() bool
	}

	// AccountDisabledError indicates that user's account is disabled.
	AccountDisabledError interface {
		UserServiceError
		IsAccountDisabled() bool
	}
)
