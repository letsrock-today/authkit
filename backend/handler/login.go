package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/letsrock-today/hydra-sample/backend/util/mapstructureutil"
)

type (
	LoginForm struct {
		Challenge string `mapstructure:"challenge" valid:"required"`
		Login     string `mapstructure:"login" valid:"email,required"`
		Password  string `mapstructure:"password" valid:"stringlength(3|10),required"`
	}
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

	var loginForm LoginForm
	if err := mapstructureutil.DecodeWithHook(
		mapstructureutil.JoinStringsFunc(""),
		r.Form,
		&loginForm); err != nil {
		writeErrorResponse(w, err)
		return
	}

	if _, err := govalidator.ValidateStruct(loginForm); err != nil {
		writeErrorResponse(w, err)
		return
	}

	log.Println("Login, loginForm:", loginForm)

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
