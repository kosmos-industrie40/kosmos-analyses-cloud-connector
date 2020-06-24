package logic

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"

	"k8s.io/klog"
)

type AnalysesInitial struct {
	Db database.Postgres
}

func (a AnalysesInitial) Analyses(db database.Postgres) {
	a.Db = db
}

// InsertResult insert a result into the database
func (a AnalysesInitial) InsertResult(contract string, machine string, sensor string, data []models.UploadResult) error {
	for _, res := range data {
		curJson, err := json.Marshal(res)
		if err != nil {
			return err
		}

		date := time.Unix(res.Date, 0)

		if err := a.Db.Insert("analyse_result", []string{"contract", "machine", "sensor", "time", "result"}, []interface{}{contract, machine, sensor, date, curJson}); err != nil {
			return err
		}
	}
	return nil
}

// GetSpecificResult returns a specific result as json
func (a AnalysesInitial) GetSpecificResult(contractId string, resultId string) ([]byte, error) {
	resId, err := strconv.ParseInt(resultId, 10, 64)
	if err != nil {
		return nil, nil
	}

	var ret []byte
	var cRet interface{} = ret
	values := []*interface{}{&cRet}

	klog.Infof("contract: %s\tid: %d", contractId, resId)
	if err := a.Db.Query("analyse_result", []string{"result"}, []string{"contract", "id"}, values, []interface{}{contractId, resId}); err != nil {
		return nil, err
	}

	return cRet.([]byte), nil
}

// GetResultSet returns an array of all results, which fulfill given parameters
func (a AnalysesInitial) GetResultSet(contractId string, queryParams map[string][]string) ([]models.ResultList, error) {
	parameters := []string{"contract"}
	parameterValue := []interface{}{contractId}
	start := time.Time{}
	end := time.Time{}

	for parameter, value := range queryParams {
		if len(value) != 1 {
			return nil, fmt.Errorf("unexpected length of the value paramter %s has %d attributes", parameter, len(value))
		}
		switch parameter {
		default:
			return nil, fmt.Errorf("unexpected query parameter found with: %s\n", parameter)
		case "machine":
			parameters = append(parameters, "machine")
			parameterValue = append(parameterValue, value[0])
		case "sensor":
			parameters = append(parameters, "sensor")
			parameterValue = append(parameterValue, value[0])
		case "start":
			parsedValue, err := strconv.ParseInt(value[0], 10, 64)
			if err != nil {
				return nil, err
			}

			start = time.Unix(parsedValue, 0)
		case "end":
			parsedValue, err := strconv.ParseInt(value[0], 10, 64)
			if err != nil {
				return nil, err
			}

			end = time.Unix(parsedValue, 0)
		}
	}

	var timeStamp []time.Time
	var machine []string
	var id []int64
	var cTime, cMachine, cId interface{}
	cTime = timeStamp
	cMachine = machine
	cId = id
	values := []*interface{}{
		&cId,
		&cTime,
		&cMachine,
	}

	klog.Infof("parameter %v\tvalues %v\t", parameters, parameterValue)
	if err := a.Db.QueryTime("analyse_result", []string{"id", "time", "machine"}, parameters, "time", start, end, values, parameterValue); err != nil {
		return nil, err
	}
	timeStamp = cTime.([]time.Time)
	machine = cMachine.([]string)
	id = cId.([]int64)

	if len(timeStamp) != len(machine) && len(machine) != len(id) {
		return nil, fmt.Errorf("the length of the array of the database result has not the same size")
	}

	var ret []models.ResultList
	for i := 0; i < len(timeStamp); i++ {
		date := timeStamp[i]
		ret = append(ret, models.ResultList{
			Id:      id[i],
			Machine: machine[i],
			Date:    date.Unix(),
		})
	}
	return ret, nil
}
