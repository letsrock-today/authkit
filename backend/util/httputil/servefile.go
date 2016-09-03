package httputil

import "net/http"

func FileServer(fname string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, fname)
	})
}
