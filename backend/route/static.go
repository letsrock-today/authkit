package route

import (
	"net/http"

	"github.com/letsrock-today/hydra-sample/backend/util/httputil"
)

func initStatic() {
	http.Handle("/", httputil.AddSuffix(".html", http.FileServer(http.Dir("../ui-web/html"))))
	http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("../ui-web/dist"))))
}
