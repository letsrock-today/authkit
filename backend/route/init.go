package route

import (
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/handler"
)

func Init() {
	initStatic()
	http.HandleFunc("/auth-providers", handler.AuthProviders)
	http.HandleFunc("/auth-code-urls", handler.AuthCodeURLs)
}
