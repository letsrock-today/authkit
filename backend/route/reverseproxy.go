package route

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/letsrock-today/hydra-sample/backend/config"
	customhttputil "github.com/letsrock-today/hydra-sample/backend/util/httputil"
)

// we use proxy for hydra requests, so that all interaction with UI went via single port

func initReverseProxy() {
	u, err := url.Parse(config.GetConfig().HydraAddr)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	//TODO: use real certeficates in PROD and remove this
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	http.Handle(
		"/hydra/",
		http.StripPrefix(
			"/hydra/",
			customhttputil.Block("*/keys/*", proxy)))
}
