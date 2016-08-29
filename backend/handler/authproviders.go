package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/config"
)

type ProvidersReply struct {
	Providers []config.OAuth2Provider `json:"providers"`
}

func AuthProviders(w http.ResponseWriter, r *http.Request) {
	p := ProvidersReply{}
	p.Providers = config.GetConfig().OAuth2Providers

	b, err := json.Marshal(p)
	if err != nil {
		log.Fatalf("Load providers: %v", err)
	}
	w.Write(b)
}
