package endpoints

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/logic"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models_database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/mqtt"
)

type MachineData struct {
	SendChan chan mqtt.Msg
	Auth     logic.Authentication
}

func (m MachineData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	// handle requests of all other http methods
	default:
		w.WriteHeader(405)
	// handle post requests
	case "POST":
		var data []models_database.Data
		var sData models_database.SendData
		var msg mqtt.Msg

		// read data from body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			klog.Errorf("could not read data from request")
			w.WriteHeader(400)
			return
		}

		// convert body to internal data type
		if err := json.Unmarshal(body, &data); err != nil {
			klog.Errorf("could not unmarshal data: %s\n", err)
			w.WriteHeader(400)
			return
		}

		// split in multiple mqtt messages
		for _, dat := range data {
			sData.Columns = dat.Columns
			sData.Data = dat.Data
			sData.From = dat.From
			sData.Machine = dat.Machine
			sData.Meta = dat.Meta
			sData.Sensor = dat.Sensor

			msg.Topic = fmt.Sprintf("kosmos/machine-data/%s/sensor/%s/update", dat.Machine, dat.Sensor) //TODO
			msg.Msg, err = json.Marshal(sData)
			if err != nil {
				klog.Errorf("could not translate to used data: %s\n", err)
				w.WriteHeader(500)
				return
			}

			// sending message to mqtt broker
			m.SendChan <- msg
		}
	}
}
