package handler

// UserService provides methods to persist users.
type UserService interface {

	// Create creates new disabled user.
	Create(login, password string) UserServiceError

	// Authenticate authenticates user, returns nil, if account exists and enabled.
	Authenticate(login, password string) UserServiceError

	// RequestEmailConfirmation requests user to confirm email address.
	RequestEmailConfirmation(login string) UserServiceError

	// return user by login
	//	User(login string) (User, UserServiceError)

	// update password
	//	UpdatePassword(login, password string) UserServiceError

	// enable account
	//	Enable(login string) UserServiceError

	// save token
	//	UpdateToken(login, pid string, token *oauth2.Token) UserServiceError

	// return token
	//	Token(login, pid string) (*oauth2.Token, UserServiceError)

	// return user by token
	// token_field is one of [accesstoken, refreshtoken]
	//	UserByToken(pid, tokenField, token string) (User, UserServiceError)
}

type User interface {
	Login() string
	Email() string
	//	PasswordHash string
	//	Disabled     *time.Time
	//	Tokens       map[string]*oauth2.Token // pid -> token
}

type UserServiceError interface {
	error
	IsDuplicateUser() bool
	IsUserNotFound() bool
	IsAccountDisabled() bool
}
