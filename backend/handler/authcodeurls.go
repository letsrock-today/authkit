package handler

import (
	"encoding/json"
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/util/jwtutil"
)

type (
	authCodeURL struct {
		Id  string `json:"id"`
		URL string `json:"url"`
	}

	authCodeURLsReply struct {
		URLs []authCodeURL `json:"urls"`
	}
)

func AuthCodeURLs(w http.ResponseWriter, r *http.Request) {
	reply := authCodeURLsReply{}
	cfg := config.GetConfig()
	for pid, conf := range cfg.OAuth2Configs {
		state, err := jwtutil.NewJWTSignedString(
			cfg.OAuth2State.TokenSignKey,
			cfg.OAuth2State.TokenIssuer,
			pid,
			cfg.OAuth2State.Expiration)
		if err != nil {
			writeErrorResponse(w, err)
			return
		}
		reply.URLs = append(reply.URLs, authCodeURL{pid, conf.AuthCodeURL(state)})
	}
	b, err := json.Marshal(reply)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
