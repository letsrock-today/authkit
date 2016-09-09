package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/util/seekingbuffer"
)

type providersReply struct {
	Providers []config.OAuth2Provider `json:"providers"`
}

func AuthProviders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Expires", time.Now().UTC().Format(http.TimeFormat))
	http.ServeContent(
		w,
		r,
		"",
		config.ModTime(),
		seekingbuffer.New(
			func() ([]byte, error) {
				p := providersReply{}
				p.Providers = config.GetConfig().OAuth2Providers

				b, err := json.Marshal(p)
				if err != nil {
					log.Println("Error at AuthProviders, json.Marshal():", err)
				}
				return b, err
			}))
}
