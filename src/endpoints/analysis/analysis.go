package analysis

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"k8s.io/klog"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/analysis/models"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/auth"
)

type Analysis interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type analysis struct {
	authHelper auth.Helper
	analysis   AnalyseLogic
}

func NewAnalysisEndpoint(analysisLogic AnalyseLogic, authHelper auth.Helper) Analysis {
	return analysis{analysis: analysisLogic, authHelper: authHelper}
}

func (a analysis) handlePost(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimRight(r.URL.Path, "/")
	ur := strings.Split(path, "/")

	if len(ur) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isAuth, statusCode, err := a.authHelper.IsAuthenticated(r, ur[2], true)
	if err != nil {
		klog.Errorf("cannot check authentication %s", err)
		w.WriteHeader(500)
		return
	}

	if !isAuth {
		w.WriteHeader(statusCode)
		return
	}

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
	var data []models.Analysis
	if err := json.Unmarshal(body, &data); err != nil {
		klog.Errorf("could not parse data: %s\n", err)
		w.WriteHeader(400)
		return
	}

	// handle request
	if err := a.analysis.InsertResult(ur[2], ur[3], ur[4], data); err != nil {
		w.WriteHeader(500)
		klog.Errorf("could not insert data: %s\n", err)
		return
	}

	w.WriteHeader(201)
}

func (a analysis) handleGet(w http.ResponseWriter, r *http.Request) {
	// removing the trailing /
	path := strings.TrimRight(r.URL.Path, "/")
	ur := strings.Split(path, "/")

	if len(ur) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isAuth, statusCode, err := a.authHelper.IsAuthenticated(r, ur[2], false)
	if err != nil {
		klog.Errorf("cannot check authentication %s", err)
		w.WriteHeader(500)
		return
	}

	klog.Infof("authentication is: %t", isAuth)

	if !isAuth {
		w.WriteHeader(statusCode)
		return
	}

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
		resSet, err := a.analysis.GetResultSet(contractId, parsedQuery)
		if err != nil {
			klog.Errorf("error occurred in GetResultSet: %v\n", err)
			w.WriteHeader(500)
			return
		}

		// if the output is empty we should not send "NULL" to the client
		if string(resSet) == "null" {
			return
		}

		// send return value
		if _, err := w.Write(resSet); err != nil {
			klog.Errorf("could not write result: %s\n", err)
			w.WriteHeader(500)
			return
		}
		return

	case 4:
		// query specific analyses

		resultId, err := strconv.ParseInt(ur[3], 10, 64)
		if err != nil {
			klog.Errorf("cannot parse result id to type")
			w.WriteHeader(400)
			return
		}

		ret, err := a.analysis.GetSpecificResult(ur[2], resultId)
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
}

func (a analysis) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	klog.Infof("receive http request on url: %s and with method: %s", r.URL.String(), r.Method)
	switch r.Method {
	// handle all http methods without get and post
	default:
		w.WriteHeader(405)
		return
	// handle get requests
	case "GET":
		a.handleGet(w, r)
	// handle post request
	case "POST":
		a.handlePost(w, r)
	}
}
