package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"

	"github.com/letsrock-today/hydra-sample/backend/service/hydra"
)

type (
	LoginForm struct {
		Challenge []string `mapstructure:"challenge" valid:"required"`
		Login     []string `mapstructure:"login" valid:"email,required"`
		Password  []string `mapstructure:"password" valid:"stringlength(3|10),required"`
	}
	LoginReply struct {
		Challenge     jwt.Claims `json:"challenge"`
		Authenticated bool       `json:"authenticated"`
		Error         string     `json:"error"`
	}
)

func Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(0); err != nil {
		log.Println(err)
	}

	// To simplify validation logic we convert map to structure first

	var loginForm LoginForm
	if err := mapstructure.Decode(r.Form, &loginForm); err != nil {
		writeErrorResponse(w, err)
		return
	}

	if _, err := govalidator.ValidateStruct(loginForm); err != nil {
		writeErrorResponse(w, err)
		return
	}

	token, err := hydra.VerifyConsentChallenge(loginForm.Challenge[0])
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	if err := UserService.Authenticate(
		loginForm.Login[0],
		loginForm.Password[0]); err != nil {
		writeErrorResponse(w, err)
		return
	}

	reply := LoginReply{
		Authenticated: true,
		Challenge:     token.Claims,
	}
	b, err := json.Marshal(reply)
	if err != nil {
		log.Fatal("Login, marshal json:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
