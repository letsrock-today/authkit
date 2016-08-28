package conrtollers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/config"
)

type ProvidersReply struct {
	Providers []config.OAuth2Provider `json:"providers"`
}

type AuthCodeURLsReplyItem struct {
	Id  string `json:"id"`
	URL string `json:"url"`
}

type AuthCodeURLsReply struct {
	URLs []AuthCodeURLsReplyItem `json:"urls"`
}

func Providers(w http.ResponseWriter, r *http.Request) {
	p := ProvidersReply{}
	p.Providers = config.GetConfig().Providers

	b, err := json.Marshal(p)
	if err != nil {
		log.Fatalf("Load providers: %#v", err)
	}
	w.Write(b)
}

func AuthCodeURLs(w http.ResponseWriter, r *http.Request) {
	reply := AuthCodeURLsReply{}
	for pid, conf := range config.GetConfig().OAuth2Configs {
		state, err := common.CreateStateToken(api.cfg.TokenSignKey, pid, r.FormValue("client"), api.cfg.OAuth2StateExpiration)
		if err != nil {
			return err
		}
		reply.URLs = append(reply.URLs, AuthCodeURLsReplyItem{pid, conf.AuthCodeURL(state, conf.GetAuthCodeOptions(r.Form)...)})
	}
	return nil
}

type customClaims struct {
	UserID string `json:"userid"`
	jwt.StandardClaims
}

func createToken(userID string) string {
	claims := customClaims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpiration).Unix(),
			Issuer:    TokenIssuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(SignKey)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func validToken(t interface{}) (string, bool) {
	if s, ok := t.(string); ok {
		token, err := jwt.ParseWithClaims(s, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
			return SignKey, nil
		})
		if err != nil {
			log.Print(err)
			return "", false
		}
		if claims, ok := token.Claims.(*customClaims); ok && token.Valid {
			return claims.UserID, true
		}
	}
	return "", false
}
