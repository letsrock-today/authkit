package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type (
	LoginReply struct {
		Challenge     string `json:"challenge"`
		Authenticated bool   `json:"authenticated"`
		Error         string `json:"error"`
	}
)

func Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(0); err != nil {
		log.Println(err)
	}
	//TODO
	log.Println("Login, params:", r.Form)
	reply := LoginReply{
		Authenticated: false,
		Error:         "Not implemented yet",
	}
	b, err := json.Marshal(reply)
	if err != nil {
		log.Fatal("Login, marshal json:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
