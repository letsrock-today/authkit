package route

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"

	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
)

// we use proxy for hydra requests, so that all interaction with UI went via single port

func initReverseProxy(e *echo.Echo) {
	c := config.Get()
	u, err := url.Parse(c.HydraAddr)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.TLSInsecureSkipVerify},
	}

	e.Any("/oauth2/*", standard.WrapHandler(proxy))
}
