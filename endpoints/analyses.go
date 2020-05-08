package endpoints

import (
	"net/http"
)

type Analyses struct{}

func (a Analyses) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(501)
}
