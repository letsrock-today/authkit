package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

var UserService userapi.UserAPI

func writeErrorResponse(w http.ResponseWriter, err error) {
	log.Println("writeErrorResponse invoked with err:", err)
	reply := struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	b, e := json.Marshal(reply)
	if e != nil {
		log.Println("writeErrorResponse, marshal json:", e, "Original error:", err)
		http.Error(w, "Error writing response.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
