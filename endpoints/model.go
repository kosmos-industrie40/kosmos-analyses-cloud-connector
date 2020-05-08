package endpoints

import "net/http"

type Model struct{}

func (m Model) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(501)
}
