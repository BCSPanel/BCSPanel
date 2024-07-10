package httpserver

import (
	"net/http"

	"github.com/bddjr/hlfhr"
)

var Server80Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	hlfhr.RedirectToHttps(w, r, 302)
})
