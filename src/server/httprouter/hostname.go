package httprouter

import "net/http"

func RequestHostname(r *http.Request) string {
	r.URL.Host = r.Host
	return r.URL.Hostname()
}

func RequestPort(r *http.Request) string {
	r.URL.Host = r.Host
	return r.URL.Port()
}
