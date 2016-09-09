package handler

import (
	"encoding/json"
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/service/hydra"
)

type (
	consentRequest struct {
		Challenge string   `json:"challenge"`
		Login     string   `json:"login"`
		Scopes    []string `json:"scopes"`
	}
	consentReply struct {
		Consent string `json:"consent"`
	}
)

func Consent(w http.ResponseWriter, r *http.Request) {
	var cr consentRequest
	err := json.NewDecoder(r.Body).Decode(&cr)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	signedTokenString, err := hydra.GenerateConsentToken(
		cr.Login,
		cr.Scopes,
		cr.Challenge)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	reply := consentReply{
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
