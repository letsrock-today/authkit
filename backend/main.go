package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/route"
)

func main() {
	flag.Parse()
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
