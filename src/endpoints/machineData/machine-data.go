package machineData

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/auth"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/mqtt"
	mqttModels "gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/mqtt/models"
)

type MachineData interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func NewMachineDataEndpoint(sendChan chan mqtt.Msg, authHelper auth.Helper, contract Contract) MachineData {
	return machineData{sendChan: sendChan, auth: authHelper, contr: contract}
}

type machineData struct {
	sendChan chan mqtt.Msg
	auth     auth.Helper
	contr Contract
}

func (m machineData) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// handle requests of all other http methods
	default:
		w.WriteHeader(405)
	// handle post requests
	case "POST":
		var data []Model
		var sData mqttModels.MachineData
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
			sData.Metadata = dat.Metadata
			sData.Timestamp = dat.Timestamp
			sData.Signature = dat.Signature

			if _, err := time.Parse(time.RFC3339, dat.Timestamp); err != nil {
				klog.Errorf("cannot validate timestamp: %s", err)
				w.WriteHeader(400)
				return
			}

			authenticated := false
			var statusCode int
			var err error
			contracts, err := m.contr.GetContracts(dat.Machine, dat.Sensor)
			for _, cont := range contracts {
				authenticated, statusCode, err = m.auth.IsAuthenticated(r, cont, true)
				if err != nil {
					klog.Errorf("cannot check authentication: %s", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if authenticated {
					break
				}
			}

			if !authenticated {
				w.WriteHeader(statusCode)
				return
			}

			msg.Topic = fmt.Sprintf("kosmos/machine-data/%s/sensor/%s/update", dat.Machine, dat.Sensor)
			msg.Msg, err = json.Marshal(sData)
			if err != nil {
				klog.Errorf("could not translate to used data: %s\n", err)
				w.WriteHeader(500)
				return
			}

			// sending message to mqtt broker
			m.sendChan <- msg
		}
	}
}
