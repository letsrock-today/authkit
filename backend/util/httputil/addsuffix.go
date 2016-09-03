package httputil

import "net/http"

func AddSuffix(suffix string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 0 && r.URL.Path != "/" {
			r.URL.Path = r.URL.Path + suffix
		}
		h.ServeHTTP(w, r)
	})
}
