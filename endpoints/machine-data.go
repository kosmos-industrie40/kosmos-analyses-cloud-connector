package endpoints

import "net/http"

type MachineData struct{}

func (m MachineData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(501)
}
