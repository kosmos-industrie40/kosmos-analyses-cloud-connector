package endpoints

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/logic"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
)

type Model struct {
	Db database.Postgres
}

func (m Model) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	if token == "" {
		w.WriteHeader(401)
		return
	}

	user, err := logic.User(token, m.Db)
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
	default:
		w.WriteHeader(405)
		return
	case "GET":
		ur := strings.Split(r.URL.Path, "/")
		if len(ur) != 3 {
			w.WriteHeader(400)
			return
		}

		data, err := logic.GetModel(ur[2], m.Db)
		if err != nil {
			w.WriteHeader(500)
			klog.Errorf("could not received update model: %s\n", err)
			return
		}

		if len(data) == 0 {
			return
		}

		printData, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(500)
			klog.Errorf("could not marshal data: %v\n", printData)
			return
		}

		if _, err := w.Write(printData); err != nil {
			w.WriteHeader(500)
			klog.Errorf("could not send data to client %s\n", err)
			return
		}

	case "PUT":

		ur := strings.Split(r.URL.Path, "/")
		if len(ur) != 3 {
			w.WriteHeader(400)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			klog.Errorf("could not read data from request body; err : %v\n", err)
			w.WriteHeader(400)
			return
		}

		var uploadedData models.UpdateModelState
		if err := json.Unmarshal(body, &uploadedData); err != nil {
			klog.Errorf("could not unmarshal data: %s\n", err)
			w.WriteHeader(400)
			return
		}

		err = logic.UpdateModel(ur[2], uploadedData, m.Db)
		if err != nil {
			w.WriteHeader(500)
			klog.Errorf("could not update model: %s\n", err)
			return
		}

		w.WriteHeader(201)
	}
}
