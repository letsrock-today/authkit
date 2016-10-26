package main

import (
	"flag"
	"html/template"
	"io"
	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"

	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/route"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/profile/mgo-profile"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user/mgo-user"
)

var (
	cfgPath, cfgName string
	debugMode        bool
)

func main() {
	flag.StringVar(&cfgPath, "cfg-dir", "", "dir with app's config")
	flag.StringVar(&cfgName, "cfg-name", "", "app's config base file name")
	flag.BoolVar(&debugMode, "dbg", false, "debug mode")
	flag.Parse()
	config.Init(cfgPath, cfgName)
	c := config.Get()

	// store implementations can be replaced here

	us, err := user.New(
		c.MongoDB().URL,
		c.MongoDB().Name,
		"users",
		c.ConfirmationLinkLifespan())
	if err != nil {
		log.Fatal(err)
	}
	defer us.Close()

	ps, err := profile.New(
		c.MongoDB().URL,
		c.MongoDB().Name,
		"profiles")
	if err != nil {
		log.Fatal(err)
	}
	defer ps.Close()

	e := echo.New()
	if debugMode {
		e.SetDebug(true)
		e.SetLogLevel(0)
	}
	e.SetHTTPErrorHandler(newHTTPErrorHandler(e.DefaultHTTPErrorHandler))

	route.Init(e, us, ps)

	e.SetRenderer(&Renderer{
		templates: template.Must(template.ParseGlob("../ui-web/templates/*.html")),
	})

	log.Printf("Serving at address: '%s'.", c.ListenAddr())
	log.Printf("Use 'https://' prefix in browser.")
	log.Printf("Press Ctrl+C to exit.")

	e.Run(standard.WithConfig(engine.Config{
		Address:     c.ListenAddr(),
		TLSCertFile: c.TLSCertFile(),
		TLSKeyFile:  c.TLSKeyFile(),
	}))
}

// Renderer used to render templates from the folder.
// In this app we use templates to render responses on email confirmation link,
// password change page, oauth2 callback, etc. At places where page designed
// with javascript frameworks seems as overkill.
type Renderer struct {
	templates *template.Template
}

func (t *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// newHTTPErrorHandler returns HTTPErrorHandler similar to default (it actually)
// wraps one) but which also outputs errors in custom format.
func newHTTPErrorHandler(prev func(error, echo.Context)) func(error, echo.Context) {
	return func(err error, c echo.Context) {
		prev(err, c)
		c.Logger().Debugf("%+v", err)
	}
}
