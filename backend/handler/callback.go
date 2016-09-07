package handler

import "net/http"

type (
	CallbackRequest struct {
		Error            string
		ErrorDescription string
	}
	CallbackReply struct {
	}
)

func Callback(w http.ResponseWriter, r *http.Request) {
	//TODO
}
