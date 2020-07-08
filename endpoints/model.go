package endpoints

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/logic"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
)

type Model struct {
	Auth  logic.Authentication
	Model logic.Model
}

func (m Model) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/2
	if token == "" {
		w.WriteHeader(401)
		return
	}

	user, err := m.Auth.User(token)
	if err != nil {
		w.WriteHeader(500)
		klog.Errorf("could not test if user is authenticated: %s", err)
		return
	}

	if user == "" {
		w.WriteHeader(401)
		return
	}

	switch r.Method {
	// handle all other http methods
	default:
		w.WriteHeader(405)
		return
	// handle get requests
	case "GET":
		// get all models, which are used in by a contract
		path := strings.TrimRight(r.URL.Path, "/")
		ur := strings.Split(path, "/")
		// wrong count of parameters
		if len(ur) != 3 {
			w.WriteHeader(400)
			return
		}

		data, err := m.Model.GetModel(ur[2])
		if err != nil {
			w.WriteHeader(500)
			klog.Errorf("could not received update model: %s\n", err)
			return
		}

		// empty response
		if len(data) == 0 {
			return
		}

		// convert data to json
		printData, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(500)
			klog.Errorf("could not marshal data: %v\n", printData)
			return
		}

		// sending data to client
		if _, err := w.Write(printData); err != nil {
			w.WriteHeader(500)
			klog.Errorf("could not send data to client %s\n", err)
			return
		}

	// handle put requests
	case "PUT":
		// update contract deployment

		path := strings.TrimRight(r.URL.Path, "/")
		ur := strings.Split(path, "/")
		// wrong count of parameters
		if len(ur) != 3 {
			w.WriteHeader(400)
			return
		}

		// read data from request
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			klog.Errorf("could not read data from request body; err : %v\n", err)
			w.WriteHeader(400)
			return
		}

		// parse data to internal data type
		var uploadedData models.UpdateModelState
		if err := json.Unmarshal(body, &uploadedData); err != nil {
			klog.Errorf("could not unmarshal data: %s\n", err)
			w.WriteHeader(400)
			return
		}

		// update model
		err = m.Model.UpdateModel(ur[2], uploadedData)
		if err != nil {
			w.WriteHeader(500)
			klog.Errorf("could not update model: %s\n", err)
			return
		}

		w.WriteHeader(201)
	}
}
