package endpoints

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/logic"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"

	"k8s.io/klog"
)

type Auth struct {
	Auth logic.Authentication
}

func (a Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		token := r.Header.Get("token")
		if token == "" {
			w.WriteHeader(204)
			return
		}
		user, err := a.Auth.User(token)
		if err != nil {
			w.WriteHeader(500)
			klog.Errorf("internal server error with database connection; err: %v", err)
		}
		if user == "" {
			w.WriteHeader(204)
			return
		}

		retUser := models.RetUser{Name: user}
		dat, err := json.Marshal(retUser)
		if err != nil {
			klog.Errorf("could not encode data to send back: %v", err)
		}
		if _, err := w.Write(dat); err != nil {
			w.WriteHeader(500)
		}

	case "POST":
		var user models.User
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			klog.Errorf("could not read data from request body; err : %v\n", err)
			w.WriteHeader(400)
			return
		}
		if err := json.Unmarshal(body, &user); err != nil {
			klog.Errorf("could not unmarshal data; error %s", err)
			w.WriteHeader(400)
			return
		}

		tok, err := a.Auth.Login(user.Name, user.Password)
		if err != nil {
			klog.Errorf("could not insert token into db: %v\n", err)
			w.WriteHeader(500)
			return
		}
		token := models.Token{Token: tok}
		sBody, err := json.Marshal(token)
		if err != nil {
			klog.Errorf("could not marshal token: %v\n", err)
		}

		if _, err := w.Write(sBody); err != nil {
			w.WriteHeader(500)
		}

	case "DELETE":
		token := r.Header.Get("token")
		if err := a.Auth.Logout(token); err != nil {
			klog.Errorf("could not delete data: %s\n", err)
			w.WriteHeader(500)
		}
		w.WriteHeader(201)
	default:
		w.WriteHeader(405)
	}
}
