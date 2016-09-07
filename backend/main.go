package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/handler"
	"github.com/letsrock-today/hydra-sample/backend/route"
	"github.com/letsrock-today/hydra-sample/backend/service/user/dummyuser"
)

func main() {
	flag.Parse()
	//TODO: replace implementation
	handler.UserService = dummyuser.New()
	route.Init()
	c := config.GetConfig()
	log.Printf("Serving at address: '%s'.", c.ListenAddr)
	log.Printf("Use 'https://' prefix in browser.")
	log.Printf("Press Ctrl+C to exit.")
	http.ListenAndServeTLS(
		c.ListenAddr,
		c.TLSCertFile,
		c.TLSKeyFile,
		nil)
}
