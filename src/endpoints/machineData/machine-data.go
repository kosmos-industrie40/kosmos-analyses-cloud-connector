package machineData

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"k8s.io/klog"

	"github.com/kosmos-industrie40/kosmos-analyses-cloud-connector/src/endpoints/auth"
	"github.com/kosmos-industrie40/kosmos-analyses-cloud-connector/src/mqtt"
	mqttModels "github.com/kosmos-industrie40/kosmos-analyses-cloud-connector/src/mqtt/models"
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
	contr    Contract
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
			var columns []mqttModels.Column
			for _, col := range dat.Body.Columns {
				column := mqttModels.Column{
					Name: col.Name,
					Type: col.Type,
					Meta: struct {
						Future      interface{} `json:"future,omitempty"`
						Unit        string      `json:"unit"`
						Description string      `json:"description"`
					}{
						Future:      col.Meta.Future,
						Unit:        col.Meta.Unit,
						Description: col.Meta.Description,
					},
				}
				columns = append(columns, column)
			}
			sData.Body.Columns = columns
			sData.Body.Data = dat.Body.Data
			sData.Body.Metadata = dat.Body.Metadata
			sData.Body.Timestamp = dat.Body.Timestamp
			sData.Signature = dat.Signature

			if _, err := time.Parse(time.RFC3339, dat.Body.Timestamp); err != nil {
				klog.Errorf("cannot validate timestamp: %s", err)
				w.WriteHeader(400)
				return
			}

			authenticated := false
			var statusCode int
			var err error
			contracts, err := m.contr.GetContracts(dat.Body.MachineID, dat.Body.Sensor)
			if err != nil {
				klog.Errorf("cannot get contract: %s", err)
				w.WriteHeader(statusCode)
				return
			}
			for _, cont := range contracts {
				authenticated, statusCode, err = m.auth.IsAuthenticated(r, cont, true)
				if err != nil {
					klog.Errorf("cannot check authentication: %s", err)
					w.WriteHeader(statusCode)
					return
				}

				if authenticated {
					break
				}
			}

			if !authenticated {
				klog.Infof("cannot authenticate: %t", authenticated)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			msg.Topic = fmt.Sprintf("kosmos/machine-data/%s/sensor/%s/update", dat.Body.MachineID, dat.Body.Sensor)
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
