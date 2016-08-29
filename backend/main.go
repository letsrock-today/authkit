package main

import (
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/controllers"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../ui-web/")
	})
	http.Handle("/dist", http.FileServer(http.Dir("../iu-web/dist")))
	http.HandleFunc("/api/v1/providers", conrtollers.Providers)
	http.HandleFunc("/api/v1/auth-code-urls", conrtollers.AuthCodeURLs)
	http.ListenAndServe(":8080", nil)
}
