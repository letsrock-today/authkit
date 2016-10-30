package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/middleware"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/socialprofile"
)

func (h handler) Friends(c echo.Context) error {
	u := c.Get(middleware.DefaultContextKey).(authkit.User)
	friends := []socialprofile.Profile{}
	// iterate over all available social networks and geather all friends
	for _, p := range h.config.OAuth2Providers {
		client := h.createHTTPClient(u, p)
		sp, err := socialprofile.New(p.ID)
		if err != nil {
			// strange, should be implemented for every network
			// skip, if not implemented
			log.Println(err)
			continue
		}
		fl, err := sp.Friends(client)
		if err != nil {
			log.Println(err)
			continue
		} else {
			friends = append(friends, fl...)
		}
	}
	return c.JSON(http.StatusOK, friends)
}
