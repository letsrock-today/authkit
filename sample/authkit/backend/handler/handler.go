package handler

import (
	"github.com/labstack/echo"
	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/profile"
)

type Handler interface {
	Profile(c echo.Context) error
	ProfileSave(c echo.Context) error
	Friends(c echo.Context) error
}

func New(
	c authkit.Config,
	ec authkit.ErrorCustomizer,
	ps profile.Service,
	cc authkit.ContextCreator) Handler {
	if c == nil || ec == nil || ps == nil {
		panic("invalid argument")
	}
	if cc == nil {
		cc = authkit.DefaultContextCreator{}
	}
	return handler{c, ec, ps, cc}
}

type handler struct {
	config          authkit.Config
	errorCustomizer authkit.ErrorCustomizer
	profiles        profile.Service
	contextCreator  authkit.ContextCreator
}
