package handler

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/config"
)

type (
	authCodeURL struct {
		Id  string `json:"id"`
		URL string `json:"url"`
	}

	authCodeURLsReply struct {
		URLs []authCodeURL `json:"urls"`
	}
)

func AuthCodeURLs(c echo.Context) error {
	reply := authCodeURLsReply{}
	cfg := config.Get()
	for pid, conf := range cfg.OAuth2Configs {
		state, err := apptoken.NewStateTokenString(
			cfg.OAuth2State.TokenIssuer,
			pid,
			cfg.OAuth2State.Expiration,
			cfg.OAuth2State.TokenSignKey)
		if err != nil {
			return err
		}
		reply.URLs = append(reply.URLs, authCodeURL{pid, conf.AuthCodeURL(state)})
	}
	return c.JSON(http.StatusOK, reply)
}
