package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

var UserService userapi.UserAPI

func writeErrorResponse(w http.ResponseWriter, err error) {
	reply := LoginReply{
		Authenticated: false,
		Error:         err.Error(),
	}
	b, err := json.Marshal(reply)
	if err != nil {
		log.Fatal("Login, marshal json:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
