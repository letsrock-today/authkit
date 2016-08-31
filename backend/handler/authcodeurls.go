package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/util/jwtutil"
)

type (
	AuthCodeURL struct {
		Id  string `json:"id"`
		URL string `json:"url"`
	}

	AuthCodeURLsReply struct {
		URLs []AuthCodeURL `json:"urls"`
	}
)

func AuthCodeURLs(w http.ResponseWriter, r *http.Request) {
	reply := AuthCodeURLsReply{}
	cfg := config.GetConfig()
	for pid, conf := range cfg.OAuth2Configs {
		state, err := jwtutil.NewJWTSignedString(
			cfg.OAuth2State.TokenSignKey,
			cfg.OAuth2State.TokenIssuer,
			pid,
			cfg.OAuth2State.Expiration)
		if err != nil {
			log.Fatal("AuthCodeURLs, create state:", err)
		}
		reply.URLs = append(reply.URLs, AuthCodeURL{pid, conf.AuthCodeURL(state)})
	}
	b, err := json.Marshal(reply)
	if err != nil {
		log.Fatal("AuthCodeURLs, marshal json:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
