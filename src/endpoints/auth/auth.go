package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/logic"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/models_database"

	"k8s.io/klog"
)

type Auth struct {
	Auth logic.Authentication
}

func (a Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// handle get request
	case "GET":
		//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/2
		token := r.Header.Get("token")
		if token == "" {
			w.WriteHeader(204)
			return
		}

		// receive specific user
		user, err := a.Auth.User(token)
		if err != nil {
			w.WriteHeader(500)
			klog.Errorf("internal server error with database connection; err: %v", err)
		}

		// if no user was found
		if user == "" {
			w.WriteHeader(204)
			return
		}

		// using data type and auto json converting
		retUser := models_database.RetUser{Name: user}
		dat, err := json.Marshal(retUser)
		if err != nil {
			klog.Errorf("could not encode data to send back: %v", err)
			w.WriteHeader(500)
		}
		if _, err := w.Write(dat); err != nil {
			w.WriteHeader(500)
		}

	// handle post requests
	case "POST":
		var user models_database.User
		// read data from the request
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			klog.Errorf("could not read data from request body; err : %v\n", err)
			w.WriteHeader(400)
			return
		}

		// parsing to internal data type
		if err := json.Unmarshal(body, &user); err != nil {
			klog.Errorf("could not unmarshal data; error %s", err)
			w.WriteHeader(400)
			return
		}

		// try to "log in"
		tok, err := a.Auth.Login(user.Name, user.Password)
		if err != nil {
			klog.Errorf("could not insert token into db: %v\n", err)
			w.WriteHeader(500)
			return
		}

		// using specific type to send back the request (using std ways to create the json)
		token := models_database.Token{Token: tok}
		sBody, err := json.Marshal(token)
		if err != nil {
			klog.Errorf("could not marshal token: %v\n", err)
		}

		// sending back the request
		if _, err := w.Write(sBody); err != nil {
			w.WriteHeader(500)
		}

	// handle delete request
	case "DELETE":
		// handle "log out"
		token := r.Header.Get("token")
		if err := a.Auth.Logout(token); err != nil {
			klog.Errorf("could not delete data: %s\n", err)
			w.WriteHeader(500)
		}
		w.WriteHeader(201)
	// handle all other http methods requests
	default:
		w.WriteHeader(405)
	}
}
