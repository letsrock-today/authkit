package handler

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

type (
	ConsentRequest struct {
		Challenge string   `json:"challenge"`
		Login     string   `json:"login"`
		Scopes    []string `json:"scopes"`
	}
	ConsentReply struct {
		Consent jwt.Token `json:"consent"`
	}
)

func Consent(w http.ResponseWriter, r *http.Request) {
	//TODO
	/*
		app.post('/api/consent', (r, w) => {
		        const {challenge, scopes, email} = r.body
		        hydra.generateConsentToken(email, scopes, challenge).then(({consent}) => {
		            w.send({consent})
		        }).catch((error) => {
		            console.log('An error occurred on consent', error)
		            w.status(500)
		            w.send(error)
		        })
		    })
	*/
}
