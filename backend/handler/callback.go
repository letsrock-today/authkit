package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/util/echo-querybinder"
)

type (
	callbackRequest struct {
		Error            string `form:"error"`
		ErrorDescription string `form:"error_description"`
		State            string `form:"state"`
		Code             string `form:"code"`
	}
)

func Callback(c echo.Context) error {
	var cr callbackRequest
	if err := querybinder.New().Bind(&cr, c); err != nil {
		return err
	}

	if cr.Error != "" {
		return fmt.Errorf("OAuth2 flow failed. Error: %s. Description: %s.", cr.Error, cr.ErrorDescription)
	}

	//TODO: validate request params

	s := fmt.Sprintf("Obtained code=%s and state=%s\n", cr.Code, cr.State)
	log.Println(s)

	//TODO
	cfg := config.Get()
	claims, err := parseStateToken(
		cfg.OAuth2State.TokenSignKey,
		cfg.OAuth2State.TokenIssuer,
		cr.State)
	if err != nil {
		return err
	}

	ss := fmt.Sprintf("Claims=%#v", claims)
	log.Println(ss)

	return c.String(http.StatusOK, s)
}
