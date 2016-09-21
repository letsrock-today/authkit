package handler

import (
	"net/http"

	"github.com/labstack/echo"
)

type (
	profileReply struct {
		//TODO
		FullName string `json:"fullname"`
	}
)

func Profile(c echo.Context) error {
	//TODO
	reply := profileReply{
		FullName: "Test Testovich",
	}
	return c.JSON(http.StatusOK, reply)
}
