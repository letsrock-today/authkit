package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/authkit/middleware"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/socialprofile"
	"github.com/letsrock-today/hydra-sample/sample/authkit/backend/service/user"
)

func (h handler) Friends(c echo.Context) error {
	u := c.Get(middleware.DefaultContextKey).(user.User)
	friends := []socialprofile.Profile{}
	// iterate over all available social networks and geather all friends
	for p := range h.config.OAuth2Providers() {
		t := u.OAuth2TokenByProviderID(p.ID())
		if t == nil {
			// no token? skip this provider
			continue
		}
		pa, err := socialprofile.New(p.ID())
		if err != nil {
			// strange, should be implemented for every network
			// skip, if not implemented
			log.Println(err)
			continue
		}
		ctx := h.contextCreator.CreateContext(p.ID())
		client := p.OAuth2Config().Client(ctx, t)
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
