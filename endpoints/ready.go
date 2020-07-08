package endpoints

import (
	"net/http"
)

type Ready struct{}

func (re Ready) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
