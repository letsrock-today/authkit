package main

import (
	"flag"
	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/handler"
	"github.com/letsrock-today/hydra-sample/backend/route"
	"github.com/letsrock-today/hydra-sample/backend/service/user/dummyuser"
)

func main() {
	flag.Parse()

	//TODO: replace implementation
	handler.UserService = dummyuser.New()

	e := echo.New()

	route.Init(e)

	c := config.GetConfig()

	log.Printf("Serving at address: '%s'.", c.ListenAddr)
	log.Printf("Use 'https://' prefix in browser.")
	log.Printf("Press Ctrl+C to exit.")

	e.Run(standard.WithConfig(engine.Config{
		Address:     c.ListenAddr,
		TLSCertFile: c.TLSCertFile,
		TLSKeyFile:  c.TLSKeyFile,
	}))
}
