package endpoints

import (
	"net/http"
)

type Contract struct{}

func (c Contract) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(501)
}
