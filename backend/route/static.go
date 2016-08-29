package route

import "net/http"

func initStatic() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../ui-web/")
	})
	http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("../ui-web/dist"))))
}
