package handler

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
)

type (
	authCodeURL struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}

	authCodeURLsReply struct {
		URLs []authCodeURL `json:"urls"`
	}
)

func (h handler) AuthCodeURLs(c echo.Context) error {
	reply := authCodeURLsReply{}
	for pid, conf := range h.config.OAuth2Configs() {
		s := h.config.OAuth2State()
		state, err := apptoken.NewStateTokenString(
			s.TokenIssuer(),
			pid,
			s.Expiration(),
			s.TokenSignKey())
		if err != nil {
			return err
		}
		reply.URLs = append(reply.URLs, authCodeURL{pid, conf.AuthCodeURL(state)})
	}
	return c.JSON(http.StatusOK, reply)
}
