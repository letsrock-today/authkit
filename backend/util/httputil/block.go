package httputil

import (
	"github.com/gobwas/glob"
	"net/http"
)

func Block(pattern string, h http.Handler) http.Handler {
	g := glob.MustCompile(pattern)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.Match(r.URL.Path) {
			http.NotFound(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
