package authkit

import "golang.org/x/oauth2"

type (

	// UserService provides methods to persist users and send confirmations.
	UserService interface {
		UserStore
		Confirmer
	}

	// UserStore provides methods to persist users.
	UserStore interface {
		tokenStore
		middlewareMethods
		handlerMethods
	}

	// Confirmer provides methods to request confirmations.
	Confirmer interface {
		// RequestEmailConfirmation requests user to confirm email address.
		RequestEmailConfirmation(login string) UserServiceError

		// RequestPasswordChangeConfirmation requests user confirmation to change password (via email).
		RequestPasswordChangeConfirmation(login, passwordHash string) UserServiceError
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
		Confirmer
	}

	tokenStore interface {
		// OAuth2Token returns OAuth2 token by login and OAuth2 provider ID.
		OAuth2Token(login, providerID string) (*oauth2.Token, UserServiceError)

		// UpdateOAuth2Token saves or updates oauth2 token for user and provider.
		UpdateOAuth2Token(login, providerID string, token *oauth2.Token) UserServiceError
	}

	middlewareMethods interface {
		// UserByAccessToken returns user data by access token.
		UserByAccessToken(providerID, accessToken string) (User, UserServiceError)

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
	}

	// User provides basic information about user, required for login logic.
	User interface {
		Login() string
		Email() string
		PasswordHash() string
	}

	// UserServiceError is a general error specific to UserService.
	// It's just an  alias for error interface.
	UserServiceError error

	causer interface {
		Cause() error
	}

	// DuplicateUserError indicates that user already exists.
	DuplicateUserError interface {
		UserServiceError
		causer
		IsDuplicateUser() bool
	}

	// UserNotFoundError indicates that user not found.
	UserNotFoundError interface {
		UserServiceError
		causer
		IsUserNotFound() bool
	}

	// AccountDisabledError indicates that user's account is disabled.
	AccountDisabledError interface {
		UserServiceError
		causer
		IsAccountDisabled() bool
	}

	// RequestConfirmationError indicates request confirmation failure.
	RequestConfirmationError interface {
		UserServiceError
		causer
		IsRequestConfirmationError() bool
	}

	userServiceError struct {
		cause error
	}

	duplicateUserError       struct{ userServiceError }
	userNotFoundError        struct{ userServiceError }
	accountDisabledError     struct{ userServiceError }
	requestConfirmationError struct{ userServiceError }
)

func (e userServiceError) Cause() error {
	return e.cause
}

// NewDuplicateUserError returns new DuplicateUserError.
func NewDuplicateUserError(cause error) DuplicateUserError {
	return duplicateUserError{userServiceError{cause}}
}

func (duplicateUserError) Error() string {
	return "duplicate user"
}

func (duplicateUserError) IsDuplicateUser() bool {
	return true
}

// NewUserNotFoundError returns new UserNotFoundError.
func NewUserNotFoundError(cause error) UserNotFoundError {
	return userNotFoundError{userServiceError{cause}}
}

func (userNotFoundError) Error() string {
	return "user not found"
}

func (userNotFoundError) IsUserNotFound() bool {
	return true
}

// NewAccountDisabledError return new AccountDisabledError.
func NewAccountDisabledError(cause error) AccountDisabledError {
	return accountDisabledError{userServiceError{cause}}
}

func (accountDisabledError) Error() string {
	return "account disabled"
}

func (accountDisabledError) IsAccountDisabled() bool {
	return true
}

// NewRequestConfirmationError returns new RequestConfirmationError.
func NewRequestConfirmationError(cause error) RequestConfirmationError {
	return requestConfirmationError{userServiceError{cause}}
}

func (requestConfirmationError) Error() string {
	return "request confirmation failed"
}

func (requestConfirmationError) IsRequestConfirmationError() bool {
	return true
}
