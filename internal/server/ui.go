package server

import (
	_ "embed"
	"net/http"
)

//go:embed ui/index.html
var indexHTML string

//go:embed ui/login.html
var loginHTML string

//go:embed ui/config.html
var configHTML string

//go:embed ui/bootstrap.html
var bootstrapHTML string

//go:embed ui/app.js
var appJS string

//go:embed ui/styles.css
var stylesCSS string

func serveHTML(doc string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(doc))
	}
}
