package handler

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/letsrock-today/authkit/authkit/apptoken"
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
	for _, p := range h.config.OAuth2Providers {
		s := h.config.OAuth2State
		state, err := apptoken.NewStateTokenString(
			s.TokenIssuer,
			p.ID,
			s.Expiration,
			s.TokenSignKey)
		if err != nil {
			return errors.WithStack(err)
		}
		reply.URLs = append(reply.URLs, authCodeURL{
			p.ID,
			p.OAuth2Config.AuthCodeURL(state)})
	}
	return c.JSON(http.StatusOK, reply)
}
