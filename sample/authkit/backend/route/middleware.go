package route

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	_middleware "github.com/letsrock-today/hydra-sample/authkit/middleware"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/hydra"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user/userapi"
)

var (
	profileMiddleware echo.MiddlewareFunc
	friendsMiddleware echo.MiddlewareFunc
)

func initMiddleware(e *echo.Echo, ua userapi.UserAPI) {
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
	}))
	e.Use(middleware.Secure())
	e.Use(middleware.CSRF())

	cfg := config.Get()
	ctx := context.WithValue(
		context.Background(),
		oauth2.HTTPClient,
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.TLSInsecureSkipVerify},
			}})
	oauth2cfg := &cfg.HydraOAuth2Config
	profileMiddleware = _middleware.AccessTokenWithConfig(
		_middleware.AccessTokenConfig{
			TokenValidator: tokenValidator{},
			UserStore:      userStore{ua},
			OAuth2Config:   oauth2cfg,
			OAuth2Context:  ctx,
		})
	friendsMiddleware = _middleware.AccessTokenWithConfig(
		_middleware.AccessTokenConfig{
			TokenValidator: tokenValidator{},
			UserStore:      userStore{ua},
			OAuth2Config:   oauth2cfg,
			OAuth2Context:  ctx,
		})
}

type tokenValidator struct{}

func (tokenValidator) Validate(token string, perm interface{}) error {
	p, ok := perm.(_middleware.DefaultPermission)
	if !ok {
		return errors.New("invalid permission object")
	}
	return hydra.ValidateAccessTokenPermissions(
		token,
		p.Resource,
		p.Action,
		p.Scopes)
}

type userStore struct {
	userapi.UserAPI
}

func (u userStore) User(token string) (interface{}, error) {
	return u.UserByToken(config.PrivPID, "accesstoken", token)
}

func (userStore) OAuth2Token(user interface{}) (*oauth2.Token, error) {
	usr, ok := user.(userapi.User)
	if !ok {
		return nil, errors.New("invalid user object")
	}
	return usr.Tokens[config.PrivPID], nil
}

func (u userStore) UpdateOAuth2Token(user interface{}, token *oauth2.Token) error {
	usr, ok := user.(userapi.User)
	if !ok {
		return errors.New("invalid user object")
	}
	return u.UpdateToken(usr.Email, config.PrivPID, token)
}

func (userStore) UserContext(user interface{}) interface{} {
	usr, ok := user.(userapi.User)
	if !ok {
		return errors.New("invalid user object")
	}
	return usr.Email
}
