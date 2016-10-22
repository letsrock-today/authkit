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

var cfgPath, cfgName string

func main() {
	flag.Parse()
	flag.StringVar(&cfgPath, "config path", "", "dir with app's config")
	flag.StringVar(&cfgName, "config name", "", "app's config base file name")
	config.Init(cfgPath, cfgName)
	cfg := config.Get()

	userCollectionName := "users"
	profileCollectionName := "profiles"

	// store implementation can be changed here
	us, err := user.New(
		cfg.MongoDB.URL,
		cfg.MongoDB.Name,
		userCollectionName,
		cfg.ConfirmationLinkLifespan)
	if err != nil {
		log.Fatal(err)
	}
	defer us.Close()

	ps, err := profile.New(
		cfg.MongoDB.URL,
		cfg.MongoDB.Name,
		profileCollectionName)
	if err != nil {
		log.Fatal(err)
	}
	defer ps.Close()

	e := echo.New()
	e.SetDebug(true)
	e.SetLogLevel(0)

	route.Init(e, us, ps)

	e.SetRenderer(&Renderer{
		templates: template.Must(template.ParseGlob("../ui-web/templates/*.html")),
	})

	c := config.Get()

	log.Printf("Serving at address: '%s'.", c.ListenAddr)
	log.Printf("Use 'https://' prefix in browser.")
	log.Printf("Press Ctrl+C to exit.")

	e.Run(standard.WithConfig(engine.Config{
		Address:     c.ListenAddr,
		TLSCertFile: c.TLSCertFile,
		TLSKeyFile:  c.TLSKeyFile,
	}))
}

//TODO: refactoring required

type Renderer struct {
	templates *template.Template
}

func (t *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
