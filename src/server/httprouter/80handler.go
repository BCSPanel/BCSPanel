package httprouter

import (
	"net/http"
	"strings"
)

var Server80Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	r.URL.Host, _ = strings.CutSuffix(r.Host, ":80")
	r.URL.Scheme = "https"
	http.Redirect(w, r, r.URL.String(), http.StatusFound)
})
