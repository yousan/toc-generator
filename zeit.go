package handler

import (
	"net/http"

	app "github.com/yousan/toc-generator/app"
)

func H(w http.ResponseWriter, r *http.Request) {
	app.Default().ServeHTTP(w, r)
}

