package route

import (
	"crypto/tls"
	"log"
	"net/http"

	"golang.org/x/oauth2"

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
	c := config.Get()

	httpclient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.TLSInsecureSkipVerify,
			}}}

	cc := authkit.NewCustomHTTPClientContextCreator(
		map[string]*http.Client{
			c.PrivateProviderID:               httpclient,
			c.PrivateProviderIDTrustedContext: httpclient,
		})

	log.Printf("########: %#v\n%#v\n", c, c.PrivateOAuth2Provider)

	as := hydra.New(
		c.HydraAddr,
		c.PrivateProviderID,
		c.PrivateProviderIDTrustedContext,
		c.ChallengeLifespan,
		c.PrivateOAuth2Provider.PrivateOAuth2Config.(*oauth2.Config),
		c.OAuth2ClientCredentials,
		c.OAuth2State.ToAuthkitType(),
		cc,
		c.TLSInsecureSkipVerify)

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
	initAPI(e, c.ToAuthkitType(), as, userService, ps, sps, cc)
}
