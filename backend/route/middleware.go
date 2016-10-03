package route

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/letsrock-today/hydra-sample/backend/config"
	_middleware "github.com/letsrock-today/hydra-sample/backend/middleware"
	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
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
	e.Use(middleware.CSRF(config.Get().CSRFSecret))

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
			PID:           config.PrivPID,
			UserAPI:       ua,
			Callback:      cbProfile,
			OAuth2Config:  oauth2cfg,
			OAuth2Context: ctx,
		})
	friendsMiddleware = _middleware.AccessTokenWithConfig(
		_middleware.AccessTokenConfig{
			PID:           config.PrivPID,
			UserAPI:       ua,
			Callback:      cbFriends,
			OAuth2Config:  oauth2cfg,
			OAuth2Context: ctx,
		})
}

func cbProfile(method, _ string) (scopes []string, resource, action string, err error) {
	switch strings.ToUpper(method) {
	case "GET":
		return []string{"core"}, "rn:api", "get", nil
	case "POST":
		return []string{"test.profile.edit"}, "rn:api:profile", "edit", nil
	}
	return []string{}, "", "", errors.New("access forbidden")
}

func cbFriends(method, uri string) (scopes []string, resource, action string, err error) {
	switch strings.ToUpper(method) {
	case "GET":
		return []string{"test.friends.view"}, "rn:api:friends", "get", nil
	}
	return []string{}, "", "", errors.New("access forbidden")
}
