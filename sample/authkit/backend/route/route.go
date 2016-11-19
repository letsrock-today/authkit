package route

import (
	"crypto/tls"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/labstack/echo"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/hydra"
	"github.com/letsrock-today/authkit/sample/authkit/backend/config"
	"github.com/letsrock-today/authkit/sample/authkit/backend/confirmer"
	"github.com/letsrock-today/authkit/sample/authkit/backend/handler"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/profile"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/socialprofile"
	"github.com/letsrock-today/authkit/sample/authkit/backend/service/user"
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

	as := hydra.New(hydra.Config{
		HydraURL:                 c.HydraAddr,
		ProviderID:               c.PrivateProviderID,
		ProviderIDTrustedContext: c.PrivateProviderIDTrustedContext,
		ChallengeLifespan:        c.ChallengeLifespan,
		OAuth2Config:             c.PrivateOAuth2Provider.PrivateOAuth2Config.(*oauth2.Config),
		ClientCredentials:        c.OAuth2ClientCredentials,
		OAuth2State:              c.OAuth2State.ToAuthkitType(),
		ContextCreator:           cc,
		TLSInsecureSkipVerify:    c.TLSInsecureSkipVerify,
	})

	sps := socialprofile.Providers()
	userService := struct {
		user.Store
		authkit.Confirmer
	}{
		us,
		confirmer.New(confirmEmailURL, confirmPasswordURL),
	}

	ac := c.ToAuthkitType()
	ac.ErrorCustomizer = handler.NewErrorCustomizer()
	ac.AuthService = as
	ac.UserService = userService
	ac.ProfileService = ps
	ac.SocialProfileServices = sps
	ac.ContextCreator = cc

	initMiddleware(e, c, as, us, cc)
	initReverseProxy(e)
	initStatic(e)
	initAPI(e, ac)
}
