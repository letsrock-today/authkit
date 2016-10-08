package handler

import (
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/config"
)

func NewHandler(c config.Config) Handler {
	return handler{c}
}

type Handler interface {
	AuthCodeURLs(echo.Context) error
	AuthProviders(echo.Context) error
}

//TODO: comment where the methods are implemented...
type handler struct {
	config config.Config
}
