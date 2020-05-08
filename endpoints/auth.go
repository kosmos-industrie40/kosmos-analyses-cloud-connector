package endpoints

import (
	"net/http"
)

type Auth struct{}

func (a Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(501)
}
