package route

import (
	"time"

	"github.com/labstack/echo"
	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/hydra"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/confirmer"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/profile"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/socialprofile"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user"
)

func Init(
	e *echo.Echo,
	us user.Store,
	ps profile.Service) {
	c := config.GetCfg()

	//TODO
	as := hydra.New(
		"",          //hydraURL
		"",          //providerID
		"",          //providerIDTrustedContext
		1*time.Hour, //challengeLifespan
		nil,         //oauth2Config
		nil,         //clientCredentials
		c.OAuth2State(),
		authkit.DefaultContextCreator{},
		false, //tlsInsecureSkipVerify
	)
	cc := authkit.DefaultContextCreator{}
	sps := socialprofile.Providers()
	userService := struct {
		user.Store
		authkit.Confirmer
	}{
		us,
		confirmer.New(confirmEmailURL, confirmPasswordURL),
	}

	initMiddleware(e, c, as, us, cc)
	initReverseProxy(e)
	initStatic(e)
	initAPI(e, c, as, userService, ps, sps, cc)
}
