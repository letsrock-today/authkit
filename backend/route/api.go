package route

import (
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/handler"
)

func initAPI() {
	http.HandleFunc("/api/auth-providers", handler.AuthProviders)
	http.HandleFunc("/api/auth-code-urls", handler.AuthCodeURLs)

	http.HandleFunc("/api/login", handler.Login)
	http.HandleFunc("/api/login-priv", handler.LoginPriv)

	http.HandleFunc("/callback", handler.Callback)
}
