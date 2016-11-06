package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

func (h handler) Logout(c echo.Context) error {
	// Get access token from header.
	req := c.Request()
	auth := req.Header().Get("Authorization")
	split := strings.SplitN(auth, " ", 2)
	if len(split) != 2 || !strings.EqualFold(split[0], "bearer") {
		return errors.WithStack(errors.New("invalid header fromat"))
	}
	token := strings.TrimSpace(split[1])
	if token == "" {
		return errors.WithStack(errors.New("invalid auth token"))
	}

	if err := h.auth.RevokeAccessToken(token); err != nil {
		return errors.WithStack(err)
	}
	if err := h.users.RevokeAccessToken(
		h.config.PrivateOAuth2Provider.ID,
		token); err != nil {
		return errors.WithStack(err)
	}

	return c.JSON(http.StatusOK, struct{}{})
}
