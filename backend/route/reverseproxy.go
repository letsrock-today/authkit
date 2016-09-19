package route

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"

	"github.com/letsrock-today/hydra-sample/backend/config"
)

// we use proxy for hydra requests, so that all interaction with UI went via single port

func initReverseProxy(e *echo.Echo) {
	u, err := url.Parse(config.Get().HydraAddr)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	//TODO: use real certeficates in PROD and remove this
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	e.Any("/oauth2/*", standard.WrapHandler(proxy))
}
