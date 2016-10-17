package handler

import "golang.org/x/oauth2"

type (
	// UserService provides methods to persist users.
	UserService interface {

		// Create creates new disabled user.
		Create(login, password string) UserServiceError

		// CreateEnabled creates new enabled user.
		CreateEnabled(login, password string) UserServiceError

		// Authenticate authenticates user, returns nil, if account exists and enabled.
		Authenticate(login, password string) UserServiceError

		// User returns user by login.
		User(login string) (User, UserServiceError)

		// UpdatePassword updates user's password.
		UpdatePassword(login, oldPasswordHash, newPassword string) UserServiceError

		// Enable enables user account.
		Enable(login string) UserServiceError

		// UpdateToken saves or updates oauth2 token for user and provider.
		UpdateToken(login, providerID string, token *oauth2.Token) UserServiceError

		// Token returns OAuth2 token by login and OAuth2 provider ID.
		Token(login, providerID string) (*oauth2.Token, UserServiceError)

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
		//	Disabled     *time.Time
		//	Tokens       map[string]*oauth2.Token // pid -> token
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
