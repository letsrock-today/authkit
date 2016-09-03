package route

import (
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/handler"
)

func Init() {
	initStatic()
	http.HandleFunc("/api/auth-providers", handler.AuthProviders)
	http.HandleFunc("/api/auth-code-urls", handler.AuthCodeURLs)
}
