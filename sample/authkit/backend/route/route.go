package route

import (
	"github.com/labstack/echo"
	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/hydra"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/confirmer"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/profile"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/socialprofile"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user"
	"golang.org/x/oauth2"
)

func Init(
	e *echo.Echo,
	us user.Store,
	ps profile.Service) {
	c := config.Get()

	as := hydra.New(
		c.HydraAddr,
		c.PrivateProviderID,
		c.PrivateProviderIDTrustedContext,
		c.ChallengeLifespan,
		c.PrivateOAuth2Provider().OAuth2Config().(*oauth2.Config),
		c.OAuth2ClientCredentials(),
		c.OAuth2State(),
		authkit.DefaultContextCreator{},
		c.TLSInsecureSkipVerify)

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
