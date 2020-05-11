package endpoints

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/logic"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"

	"k8s.io/klog"
)

type Contract struct{
	Db database.Postgres
}

func (c Contract) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	user, err := logic.User(token, c.Db)
	if err != nil {
		w.WriteHeader(500)
		klog.Errorf("could not test if user is authenticated")
		return
	}

	if user == "" {
		w.WriteHeader(401)
		return
	}

	switch r.Method {
	case "GET":
		splitted := strings.Split(r.URL.Path, "/")
		switch len(splitted) {
		default:
			w.WriteHeader(400)
		case 2:
			contracts, err := logic.GetAllContracts(c.Db)
			if err != nil {
				klog.Errorf("could not query all contracts: %s\n", err)
				w.WriteHeader(500)
				return
			}
			res, err := json.Marshal(contracts)
			if err != nil {
				klog.Errorf("could not marshal data: %v\n", err)
				w.WriteHeader(500)
				return
			}

			if _, err := w.Write(res); err != nil {
				w.WriteHeader(500)
				klog.Errorf("could not send message %v\n", err)
				return
			}
			w.WriteHeader(200)
		case 3:
			if splitted[2] == "" {
				contracts, err := logic.GetAllContracts(c.Db)
				if err != nil {
					klog.Errorf("could not query all contracts: %s\n", err)
					w.WriteHeader(500)
					return
				}
				res, err := json.Marshal(contracts)
				if err != nil {
					klog.Errorf("could not marshal data: %v\n", err)
					w.WriteHeader(500)
					return
				}

				if _, err := w.Write(res); err != nil {
					w.WriteHeader(500)
					klog.Errorf("could not send message %v\n", err)
					return
				}
				w.WriteHeader(200)
			} else {
				contractId := splitted[2]
				data, err := logic.GetContract(contractId, c.Db)
				if err != nil {
					klog.Errorf("could not receive contract: %s\n", err)
					w.WriteHeader(500)
					return
				}
				res, err := json.Marshal(data)
				if err != nil {
					klog.Errorf("could not marshal contract: %v\n", err)
					w.WriteHeader(500)
					return
				}
				if _, err := w.Write(res); err != nil {
					klog.Errorf("could not return result: %v\n", err)
					w.WriteHeader(500)
					return
				}
				w.WriteHeader(200)
			}
		}
	case "DELETE":
		splitted := strings.Split(r.URL.Path, "/")
		if len(splitted) != 3 {
			klog.Infof("wrong count of parameters")
			w.WriteHeader(400)
		}
		if err := logic.DeleteContract(splitted[2], c.Db); err != nil {
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(201)
	case "POST":
		var contract models.Contract

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			klog.Errorf("could not read data from request: %s", err)
			w.WriteHeader(500)
			return
		}

		err = json.Unmarshal(body, &contract)
		if err != nil {
			klog.Errorf("could not parse query parameter: %s", err)
			w.WriteHeader(400)
			return
		}

		if err := logic.InsertContract(contract, c.Db); err != nil {
			klog.Errorf("could not insert data into db: %s\n", err)
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(201)

	default:
		w.WriteHeader(405)
	}
}
