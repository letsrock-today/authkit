package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

type (
	callbackRequest struct {
		Error            []string `mapstructure:"error"`
		ErrorDescription []string `mapstructure:"error_description"`
		State            []string `mapstructure:"state"`
		Code             []string `mapstructure: code`
	}
)

func Callback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		http.Error(w, "Error decoding request params.", http.StatusInternalServerError)
		return
	}

	var cr callbackRequest
	if err := mapstructure.Decode(r.Form, &cr); err != nil {
		log.Println(err)
		http.Error(w, "Error decoding request params.", http.StatusInternalServerError)
		return
	}

	if len(cr.Error) > 0 {
		msg := fmt.Sprintf("OAuth2 flow failed. Error: %s. Description: %s.", cr.Error, cr.ErrorDescription)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	//TODO: validate request params

	log.Println("Obtained code and state", cr.Code, cr.State)

	//TODO
}
