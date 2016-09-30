package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/service/socialprofile"
)

func Friends(c echo.Context) error {
	login := c.Get("user-login").(string)
	u, err := Users.User(login)
	if err != nil {
		return err
	}
	cfg := config.Get()
	friends := []socialprofile.Profile{}
	// iterate over all available social networks and geather all friends
	for pid, oauth2cfg := range cfg.OAuth2Configs {
		token, ok := u.Tokens[pid]
		if !ok {
			// no token? skip this provider
			continue
		}
		pa, err := socialprofile.New(pid)
		if err != nil {
			// strange, should be implemented for every network
			// skip, if not implemented
			log.Println(err)
			continue
		}
		ctx := context.Background()
		client := oauth2cfg.Client(ctx, token)
		fl, err := pa.Friends(client)
		if err != nil {
			// method not implemented, or other error - just skip it
			log.Println(err)
			continue
		}
		friends = append(friends, fl...)
	}
	return c.JSON(http.StatusOK, friends)
}
