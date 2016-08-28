package conrtollers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/common"
	"github.com/letsrock-today/hydra-sample/backend/config"
)

type (
	ProvidersReply struct {
		Providers []config.OAuth2Provider `json:"providers"`
	}

	AuthCodeURLsReplyItem struct {
		Id  string `json:"id"`
		URL string `json:"url"`
	}

	AuthCodeURLsReply struct {
		URLs []AuthCodeURLsReplyItem `json:"urls"`
	}
)

func Providers(w http.ResponseWriter, r *http.Request) {
	p := ProvidersReply{}
	p.Providers = config.GetConfig().OAuth2Providers

	b, err := json.Marshal(p)
	if err != nil {
		log.Fatalf("Load providers: %#v", err)
	}
	w.Write(b)
}

func AuthCodeURLs(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("In AuthCodeURLs...\n")
	reply := AuthCodeURLsReply{}
	cfg := config.GetConfig()
	for pid, conf := range cfg.OAuth2Configs {
		fmt.Printf("In AuthCodeURLs pid=%s...\n", pid)
		state := common.CreateToken(cfg.AuthConfig.TokenSignKey, cfg.AuthConfig.TokenIssuer, pid, cfg.AuthConfig.OAuth2StateExpiration)
		reply.URLs = append(reply.URLs, AuthCodeURLsReplyItem{pid, conf.AuthCodeURL(state, conf.GetAuthCodeOptions(r.Form)...)})
	}
	b, err := json.Marshal(reply)
	if err != nil {
		log.Fatalf("Load auth code urls: %#v", err)
	}
	w.Write(b)
}
