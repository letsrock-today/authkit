package route

import (
	"fmt"

	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
	"github.com/letsrock-today/hydra-sample/authkit/handler"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	_handler "github.com/letsrock-today/hydra-sample/sample/authkit/backend/handler"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/hydra"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/profile/profileapi"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/socialprofile"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user/userapi"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/util/email"
)

func initAPI(e *echo.Echo, ua userapi.UserAPI, pa profileapi.ProfileAPI) {

	c := config.GetCfg()
	h := handler.NewHandler(c, ec{}, as{}, us{ua}, ps{pa}, sps{}, cc{})

	e.GET("/api/auth-providers", h.AuthProviders)
	e.GET("/api/auth-code-urls", h.AuthCodeURLs)

	e.GET("/api/profile", _handler.Profile, profileMiddleware)
	e.POST("/api/profile", _handler.ProfileSave, profileMiddleware)
	e.GET("/api/friends", _handler.Friends, friendsMiddleware)

	e.POST("/api/login", h.ConsentLogin)
	e.POST("/api/login-priv", h.Login)

	e.GET("/callback", h.Callback)

	e.POST("/password-reset", h.RestorePassword)
	e.POST("/password-change", h.ChangePassword)

	e.GET("/email-confirm", h.ConfirmEmail)
}

//TODO: refactoring required

type jsonError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ec struct{}

func (ec) InvalidRequestParameterError(e error) interface{} {
	return jsonError{"invalid_req_param", e.Error()}
}

func (ec) UserCreationError(e error) interface{} {
	switch e.(type) {
	case authkit.AccountDisabledError:
		return jsonError{"account_disabled", e.Error()}
	case authkit.DuplicateUserError:
		return jsonError{"duplicate_account", e.Error()}
	}
	return jsonError{"unknown_err", e.Error()}
}

func (ec) UserAuthenticationError(e error) interface{} {
	return jsonError{"auth_err", e.Error()}
}

type as struct{}

func (as) GenerateConsentToken(
	subj string,
	scopes []string,
	challenge string) (string, error) {
	return hydra.GenerateConsentToken(subj, scopes, challenge)
}

func (as) IssueConsentToken(
	clientID string,
	scopes []string) (string, error) {
	return hydra.IssueConsentToken(clientID, scopes)
}

type us struct {
	userapi.UserAPI
}

func (us us) Create(login, password string) authkit.UserServiceError {
	err := us.UserAPI.Create(login, password)
	if err != nil {
		return userServiceError{err}
	}
	return nil
}

func (us us) Authenticate(login, password string) authkit.UserServiceError {
	err := us.UserAPI.Authenticate(login, password)
	if err != nil {
		return userServiceError{err}
	}
	return nil
}

func (us us) User(login string) (authkit.User, authkit.UserServiceError) {
	u, err := us.UserAPI.User(login)
	if err != nil {
		return nil, userServiceError{err}
	}
	return user{u}, nil
}

func (us us) UpdatePassword(login, oldPasswordHash, newPassword string) authkit.UserServiceError {
	u, err := us.UserAPI.User(login)
	if err != nil {
		return userServiceError{err}
	}
	//TODO: move this into store in one request to DB
	if u.PasswordHash != oldPasswordHash {
		return userServiceError{userapi.AuthError}
	}
	err = us.UserAPI.UpdatePassword(login, newPassword)
	if err != nil {
		return userServiceError{err}
	}
	return nil
}

func (us us) Enable(login string) authkit.UserServiceError {
	err := us.UserAPI.Enable(login)
	if err != nil {
		return userServiceError{err}
	}
	return nil
}

func (us us) RequestEmailConfirmation(login string) authkit.UserServiceError {
	err := sendConfirmationEmail(login, "", confirmEmailURL, false)
	if err != nil {
		return userServiceError{err}
	}
	return nil
}

func (us us) RequestPasswordChangeConfirmation(login, passwordHash string) authkit.UserServiceError {
	err := sendConfirmationEmail(login, passwordHash, confirmPasswordURL, false)
	if err != nil {
		return userServiceError{err}
	}
	return nil
}

type user struct {
	user *userapi.User
}

func (u user) Login() string {
	return u.user.Email
}

func (u user) Email() string {
	return u.user.Email
}

func (u user) PasswordHash() string {
	return u.user.PasswordHash
}

type userServiceError struct {
	e error
}

func (e userServiceError) Error() string {
	return e.e.Error()
}

func (e userServiceError) IsDuplicateUser() bool {
	return e.e == userapi.AuthErrorDup
}

func (e userServiceError) IsUserNotFound() bool {
	return e.e == userapi.AuthErrorUserNotFound
}

func (e userServiceError) IsAccountDisabled() bool {
	return e.e == userapi.AuthErrorDisabled
}

//TODO: this const is better stay in this package, but refactor
const confirmEmailURL = "/email-confirm"
const confirmPasswordURL = "/password-confirm"

//TODO: it's temporary, remove it completely from this package
func sendConfirmationEmail(
	to, passwordhash string,
	urlpath string,
	resetPassword bool) error {
	cfg := config.Get()
	token, err := apptoken.NewEmailTokenString(
		cfg.OAuth2State.TokenIssuer,
		to,
		passwordhash,
		cfg.ConfirmationLinkLifespan,
		cfg.OAuth2State.TokenSignKey)
	if err != nil {
		return err
	}

	externalURL := cfg.ExternalBaseURL + urlpath
	link := fmt.Sprintf("%s?token=%s", externalURL, token)
	var text, topic string
	if resetPassword {
		text = fmt.Sprintf("Follow this link to change your password: %s\n", link)
		topic = "Confirm password reset"
	} else {
		text = fmt.Sprintf("Follow this link to confirm your email address and complete creating account: %s\n", link)
		topic = "Confirm account creation"
	}
	return email.Send(to, topic, text)
}

type ps struct {
	ps profileapi.ProfileAPI
}

func (ps ps) EnsureExists(login string) error {
	return ps.ps.Save(login, &socialprofile.Profile{Email: login})
}
