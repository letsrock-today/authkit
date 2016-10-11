package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/hydra-sample/authkit/seekingbuffer"
)

type (
	providersReply struct {
		Providers []oauth2Provider `json:"providers"`
	}

	oauth2Provider struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		IconURL string `json:"iconUrl"`
	}
)

func (h handler) AuthProviders(c echo.Context) error {
	c.Response().Header().Set("Expires", time.Now().UTC().Format(http.TimeFormat))
	c.ServeContent(
		seekingbuffer.New(
			func() ([]byte, error) {
				p := providersReply{}
				for _, pp := range h.config.OAuth2Providers() {
					p.Providers = append(p.Providers, oauth2Provider{
						ID:      pp.ID(),
						Name:    pp.Name(),
						IconURL: pp.IconURL(),
					})
				}

				b, err := json.Marshal(p)
				if err != nil {
					c.Logger().Error(errors.WithStack(err))
				}
				return b, err
			}),
		".json", // mapped to correct type via /etc/mime.types (if not, register it manually)
		h.config.ModTime())
	return nil
}
