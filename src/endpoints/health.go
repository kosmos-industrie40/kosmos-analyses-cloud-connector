package endpoints

import (
	"net/http"
)

type Health struct{}

func (h Health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
