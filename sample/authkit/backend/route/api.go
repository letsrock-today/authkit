package route

import (
	"fmt"

	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
	"github.com/letsrock-today/hydra-sample/authkit/handler"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	_handler "github.com/letsrock-today/hydra-sample/sample/authkit/backend/handler"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/hydra"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user/userapi"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/util/email"
)

func initAPI(e *echo.Echo, ua userapi.UserAPI) {

	c := config.GetCfg()
	h := handler.NewHandler(c, ec{}, as{}, us{ua})

	e.GET("/api/auth-providers", h.AuthProviders)
	e.GET("/api/auth-code-urls", h.AuthCodeURLs)

	e.GET("/api/profile", _handler.Profile, profileMiddleware)
	e.POST("/api/profile", _handler.ProfileSave, profileMiddleware)
	e.GET("/api/friends", _handler.Friends, friendsMiddleware)

	e.POST("/api/login", h.ConsentLogin)
	e.POST("/api/login-priv", h.Login)

	e.GET("/callback", _handler.Callback)

	e.POST("/password-reset", _handler.ResetPassword)
	e.POST("/password-change", _handler.ChangePassword)

	e.GET("/email-confirm", _handler.EmailConfirm)
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
	err, ok := e.(handler.UserServiceError)
	if ok {
		if err.IsAccountDisabled() {
			return jsonError{"account_disabled", e.Error()}
		}
		if err.IsDuplicateUser() {
			return jsonError{"duplicate_account", e.Error()}

		}
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

func (us us) Create(login, password string) handler.UserServiceError {
	err := us.UserAPI.Create(login, password)
	return userServiceError{err}
}

func (us us) Authenticate(login, password string) handler.UserServiceError {
	err := us.UserAPI.Authenticate(login, password)
	return userServiceError{err}
}

func (us us) RequestEmailConfirmation(login string) handler.UserServiceError {
	err := sendConfirmationEmail(login, "", confirmEmailURL, false)
	return userServiceError{err}
}

type userServiceError struct {
	error
}

func (e userServiceError) IsDuplicateUser() bool {
	return e == userapi.AuthErrorDup
}

func (e userServiceError) IsUserNotFound() bool {
	return e == userapi.AuthErrorUserNotFound
}

func (e userServiceError) IsAccountDisabled() bool {
	return e == userapi.AuthErrorDisabled
}

//TODO: this const is better stay in this package, but refactor
const confirmEmailURL = "/email-confirm"

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
