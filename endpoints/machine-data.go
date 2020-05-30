package endpoints

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/logic"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/mqtt"
)

type MachineData struct {
	SendChan chan mqtt.Msg
	Auth     logic.Authentication
}

func (m MachineData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

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
	default:
		w.WriteHeader(405)
	case "POST":
		var data []models.Data
		var sData models.SendData
		var msg mqtt.Msg

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			klog.Errorf("could not read data from request")
			w.WriteHeader(400)
			return
		}

		if err := json.Unmarshal(body, &data); err != nil {
			klog.Errorf("could not unmarshal data: %s\n", err)
			w.WriteHeader(400)
			return
		}

		for _, dat := range data {
			sData.Columns = dat.Columns
			sData.Data = dat.Data
			sData.From = dat.From
			sData.Machine = dat.Machine
			sData.Meta = dat.Meta
			sData.Sensor = dat.Sensor

			msg.Topic = fmt.Sprintf("kosmos/machine-data/%s/sensor/%s", dat.Machine, dat.Sensor) //TODO
			msg.Msg, err = json.Marshal(sData)
			if err != nil {
				klog.Errorf("could not translate to used data: %s\n", err)
				w.WriteHeader(500)
				return
			}

			m.SendChan <- msg
		}
	}
}
