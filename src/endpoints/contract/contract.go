package contract

import (
	"io/ioutil"
	"net/http"
	"strings"

	"k8s.io/klog"

	"github.com/kosmos-industrie40/kosmos-analyses-cloud-connector/src/endpoints/auth"
)

type Contract interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

func NewContractEndpoint(contractLogic Logic, auth auth.Helper) Contract {
	return contract{contract: contractLogic, auth: auth}
}

type contract struct {
	contract Logic
	auth     auth.Helper
}

func (c contract) handleGet(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimRight(r.URL.Path, "/")
	split := strings.Split(path, "/")
	switch len(split) {
	// not enough or to many parameter are transmitted
	default:
		w.WriteHeader(400)
		return
	// query all contracts
	case 2:
		cAuth, err := c.auth.TokenValid(r)
		if err != nil {
			klog.Errorf("cannot check token validation: %split", err)
		}

		if !cAuth {
			w.WriteHeader(http.StatusBadRequest)
			klog.Errorf("check validation returned false")
			return
		}

		contracts, err := c.contract.GetAllContracts(r.Header.Get("token"))
		if err != nil {
			klog.Errorf("could not query all contracts: %split\n", err)
			w.WriteHeader(500)
			return
		}

		if len(contracts) == 0 {
			return
		}

		if _, err := w.Write(contracts); err != nil {
			w.WriteHeader(500)
			klog.Errorf("could not send message %v\n", err)
			return
		}
	// query specific contract
	case 3:
		contractId := split[2]
		valid, statusCode, err := c.auth.IsAuthenticated(r, contractId, false)
		if err != nil {
			w.WriteHeader(statusCode)
			klog.Errorf("cannot check authentication: %s", err)
			return
		}

		if !valid {
			w.WriteHeader(http.StatusUnauthorized)
			klog.Errorf("authentication is not valid")
			return
		}

		data, err := c.contract.GetContract(contractId)
		if err != nil {
			klog.Errorf("could not receive contract: %split\n", err)
			w.WriteHeader(500)
			return
		}
		if _, err := w.Write(data); err != nil {
			klog.Errorf("could not return result: %v\n", err)
			w.WriteHeader(500)
			return
		}
	}
}

func (c contract) handleDelete(w http.ResponseWriter, r *http.Request) {
	hasRight, responseCode, err := c.auth.ContractWriteAccess(r)
	if err != nil {
		w.WriteHeader(responseCode)
		return
	}
	if !hasRight {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	split := strings.Split(r.URL.Path, "/")
	// test if the correct count of parameters has been transmitted
	if len(split) != 3 {
		klog.Infof("wrong count of parameters")
		w.WriteHeader(400)
		return
	}
	// delete contract
	if err := c.contract.DeleteContract(split[2]); err != nil {
		klog.Errorf("could not update contract: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c contract) handlePost(w http.ResponseWriter, r *http.Request) {

	hasRight, responseCode, err := c.auth.ContractWriteAccess(r)
	if err != nil {
		w.WriteHeader(responseCode)
		return
	}
	if !hasRight {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// read data from body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("could not read data from request: %s", err)
		w.WriteHeader(500)
		return
	}

	state, err := c.contract.InsertContract(body)
	if err != nil {
		klog.Errorf("could not insert data into db: %s\n", err)
	}

	w.WriteHeader(state)

}

func (c contract) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// handle get requests
	case "GET":
		c.handleGet(w, r)
	// handle delete requests
	case "DELETE":
		c.handleDelete(w, r)
	// handle post request
	case "POST":
		c.handlePost(w, r)
	// handle all other http method requests
	default:
		w.WriteHeader(405)
	}
}
