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
	//TODO to obtain user we would use custom middleware, which will get user by access token for PrivPID
	reply := profileReply{
		FullName: "Test Testovich",
	}
	return c.JSON(http.StatusOK, reply)
}
