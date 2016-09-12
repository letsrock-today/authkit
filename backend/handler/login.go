package handler

import (
	"encoding/json"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/mitchellh/mapstructure"

	"github.com/letsrock-today/hydra-sample/backend/service/hydra"
)

type (
	loginForm struct {
		Challenge []string `mapstructure:"challenge" valid:"required"`
		Login     []string `mapstructure:"login" valid:"email,required"`
		Password  []string `mapstructure:"password" valid:"stringlength(3|10),required"`
		Scopes    []string `mapstructure:"scopes" valid:"stringlength(1|500),required"`
	}
	loginReply struct {
		Consent string `json:"consent"`
	}
)

func Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(0); err != nil {
		writeErrorResponse(w, err)
		return
	}

	// To simplify validation logic we convert map to structure first

	var lf loginForm
	if err := mapstructure.Decode(r.Form, &lf); err != nil {
		writeErrorResponse(w, err)
		return
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		writeErrorResponse(w, err)
		return
	}

	if err := UserService.Authenticate(
		lf.Login[0],
		lf.Password[0]); err != nil {
		writeErrorResponse(w, err)
		return
	}

	signedTokenString, err := hydra.GenerateConsentToken(
		lf.Login[0],
		lf.Scopes,
		lf.Challenge[0])
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	reply := loginReply{
		Consent: signedTokenString,
	}
	b, err := json.Marshal(reply)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
