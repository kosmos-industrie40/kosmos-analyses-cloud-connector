package contract

import (
	"encoding/json"
	"net/http"

	"k8s.io/klog"

	"github.com/kosmos-industrie40/kosmos-analyses-cloud-connector/src/endpoints/contract/models"
)

type Logic interface {
	// GetAllContracts
	GetAllContracts(string) ([]byte, error)

	// GetContract
	GetContract(string) ([]byte, error)

	// DeleteContract
	DeleteContract(string) error

	// InsertContract
	InsertContract([]byte) (int, error)
}

type logic struct {
	resultList models.ResultList
	handler    models.ContractHandler
	system     string
}

func (c logic) GetContract(contract string) ([]byte, error) {
	con, err := c.handler.GetContract(contract)
	klog.Info(con)
	if err != nil {
		return nil, err
	}

	return json.Marshal(con)
}

func (c logic) DeleteContract(contract string) error {
	return c.handler.DeleteContract(contract)
}

func (c logic) InsertContract(bytes []byte) (int, error) {
	var contract models.Contract
	if err := json.Unmarshal(bytes, &contract); err != nil {
		klog.Infof("contract cannot be parsed: %s, received data: %s", err, string(bytes))
		return http.StatusBadRequest, err
	}

	if !contract.Valid(c.system) {
		klog.Infof("contract is not valid")
		return http.StatusBadRequest, nil
	}

	if err := c.handler.InsertContract(contract); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func (c logic) GetAllContracts(token string) ([]byte, error) {
	ids, err := c.resultList.GetAllContracts(token)
	if err != nil {
		return nil, err
	}

	return json.Marshal(ids)
}

func NewContractLogic(list models.ResultList, handler models.ContractHandler, system string) Logic {
	return logic{resultList: list, handler: handler, system: system}
}
