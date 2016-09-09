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
	}
	loginReply struct {
		Challenge     hydra.ChallengeClaims `json:"challenge"`
		Authenticated bool                  `json:"authenticated"`
	}
)

func Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(0); err != nil {
		writeErrorResponse(w, err)
		return
	}

	// To simplify validation logic we convert map to structure first

	var loginForm loginForm
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

	reply := loginReply{
		Authenticated: true,
		Challenge:     *token.Claims.(*hydra.ChallengeClaims),
	}
	b, err := json.Marshal(reply)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
