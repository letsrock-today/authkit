package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo"

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

	s := fmt.Sprintf("Obtained code=%s and state=%s", cr.Code, cr.State)

	//TODO

	log.Println(s)

	return c.String(http.StatusOK, s)
}
