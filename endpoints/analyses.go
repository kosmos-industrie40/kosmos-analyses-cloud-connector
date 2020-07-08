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

type Analyses struct {
	Auth     logic.Authentication
	Analyses logic.Analyses
}

func (a Analyses) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/2
	if token == "" {
		w.WriteHeader(401)
		return
	}

	user, err := a.Auth.User(token)
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
	// handle all http methods without get and post
	default:
		w.WriteHeader(405)
		return
	// handle get requests
	case "GET":

		// removing the trailing /
		path := strings.TrimRight(r.URL.Path, "/")
		ur := strings.Split(path, "/")
		switch len(ur) {
		// not enough parameter are transmitted through the URL
		default:
			klog.Errorf("unexpected count of parameters in query: %d\n", len(ur))
			w.WriteHeader(400)
			return
		// query all analyses
		case 3:
			parsedQuery := make(map[string][]string)
			contractId := ur[2]
			queryParam := r.URL.Query()
			for i, v := range queryParam {
				parsedQuery[i] = v
			}

			// receive the result, which should be send to the client
			resSet, err := a.Analyses.GetResultSet(contractId, parsedQuery)
			if err != nil {
				klog.Errorf("could not query result set: %v\n", err)
				w.WriteHeader(500)
				return
			}

			// convert data to json string
			retValue, err := json.Marshal(resSet)
			if err != nil {
				klog.Errorf("could convert ResultSet to json: %v\n", err)
				w.WriteHeader(500)
				return
			}

			// if the output is empty we should not send "NULL" to the client
			if string(retValue) == "null" {
				return
			}

			// send return value
			if _, err := w.Write(retValue); err != nil {
				klog.Errorf("could not write result: %s\n", err)
				w.WriteHeader(500)
				return
			}
			return

		case 4:
			// query specific analyses
			ret, err := a.Analyses.GetSpecificResult(ur[2], ur[3])
			if err != nil {
				klog.Errorf("could not query specific result: %s\n", err)
				w.WriteHeader(500)
			}
			// sending result
			if _, err := w.Write(ret); err != nil {
				klog.Errorf("could send result: %s\n", err)
				w.WriteHeader(500)
			}
		}
	// handle post request
	case "POST":
		path := strings.TrimRight(r.URL.Path, "/")
		ur := strings.Split(path, "/")
		// not enough parameter are transmitted
		if len(ur) != 5 {
			klog.Errorf("unexpected length of url path: %d\n", len(ur))
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

		// convert json to used data types
		var data []models.UploadResult
		if err := json.Unmarshal(body, &data); err != nil {
			klog.Errorf("could not parse data: %s\n", err)
			w.WriteHeader(400)
			return
		}

		// handle request
		if err := a.Analyses.InsertResult(ur[2], ur[3], ur[4], data); err != nil {
			w.WriteHeader(500)
			klog.Errorf("could not insert data: %s\n", err)
			return
		}

		w.WriteHeader(201)
		return
	}
}
