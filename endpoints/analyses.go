package endpoints

import (
	"net/http"
	"strings"
	"encoding/json"

	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/logic"
)

type Analyses struct {
	Db database.Postgres
}

func (a Analyses) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	if token == "" {
		w.WriteHeader(401)
		return
	}

	user, err := logic.User(token, a.Db)
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
		switch len(ur) {
		default:
			klog.Errorf("unexpected count of parameters in query: %d\n", len(ur))
			w.WriteHeader(400)
			return
		case 2:
			parsedQuery := make(map[string][]string)
			contractId := ur[2]
			queryParam := r.URL.Query()
			for i, v := range queryParam {
				parsedQuery[i] = v
			}
			resSet, err := logic.GetResultSet(contractId, parsedQuery, a.Db)
			if err != nil {
				klog.Errorf("could not query result set: %v\n", err)
				w.WriteHeader(500)
				return
			}

			retValue, err := json.Marshal(resSet)
			if err != nil {
				klog.Errorf("could convert ResultSet to json: %v\n", err)
				w.WriteHeader(500)
				return
			}

			if _, err := w.Write(retValue); err != nil {
				klog.Errorf("could not write result: %s\n", err)
				w.WriteHeader(500)
				return
			}
			return

		case 3:
			if ur[3] == "" {
				// same as len(ur) == 2
				parsedQuery := make(map[string][]string)
				contractId := ur[2]
				queryParam := r.URL.Query()
				for i, v := range queryParam {
					parsedQuery[i] = v
				}
				resSet, err := logic.GetResultSet(contractId, parsedQuery, a.Db)
				if err != nil {
					klog.Errorf("could not query result set: %v\n", err)
					w.WriteHeader(500)
					return
				}

				retValue, err := json.Marshal(resSet)
				if err != nil {
					klog.Errorf("could convert ResultSet to json: %v\n", err)
					w.WriteHeader(500)
					return
				}

				if _, err := w.Write(retValue); err != nil {
					klog.Errorf("could not write result: %s\n", err)
					w.WriteHeader(500)
					return
				}
				return
			} else {
				ret, err := logic.GetSpecificResult(ur[2], ur[3], a.Db)
				if err != nil {
					klog.Errorf("could not query specific result: %s\n", err)
				}
				if _, err := w.Write(ret); err != nil {
					klog.Errorf("could send result: %s\n", err)
					w.WriteHeader(500)
				}
			}
		}
	case "POST":
	}
}
