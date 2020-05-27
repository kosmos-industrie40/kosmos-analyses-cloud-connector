package logic

import (
	"encoding/json"
	"fmt"
	"strconv"
	tim "time"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
)

// InsertResult insert a result into the database
func InsertResult(contract string, machine string, sensor string, jsonData []byte, db database.Postgres) error {
	var date models.UploadResult
	if err := json.Unmarshal(jsonData, &date); err != nil {
		return err
	}

	err := db.Insert("analyse_result", []string{"contract, machine, sensor, date, json"}, []interface{}{contract, machine, sensor, date, jsonData})
	return err
}

// GetSpecificResult returns a specific result as json
func GetSpecificResult(contractId string, resultId string, db database.Postgres) ([]byte, error) {
	resId, err := strconv.ParseInt(resultId, 10, 64)
	if err != nil {
		return nil, nil
	}

	var ret []byte
	var cRet interface{} = ret
	values := []*interface{}{&cRet}

	if err := db.Query("analyse-results", []string{"result"}, []string{"id", "machine"}, values, []interface{}{contractId, resId}); err != nil {
		return nil, err
	}

	return cRet.([]byte), nil
}

// GetResultSet returns an array of all results, which fulfill given parameters
func GetResultSet(contractId string, queryParams map[string][]string, db database.Postgres) ([]models.ResultList, error) {
	parameters := []string{"contract"}
	parameterValue := []interface{}{contractId}
	start := tim.Time{}
	end := tim.Time{}

	for parameter, value := range queryParams {
		switch parameter {
		default:
			return nil, fmt.Errorf("unexpected query parameter found with: %s\n", parameter)
		case "machine":
			if len(value) != 1 {
				return nil, fmt.Errorf("unexpected length of the value paramter %s has %d attributes", parameter, len(value))
			}
			parameters = append(parameters, "machine")
			parameterValue = append(parameterValue, value[0])
		case "sensor":
			if len(value) != 1 {
				return nil, fmt.Errorf("unexpected length of the value paramter %s has %d attributes", parameter, len(value))
			}
			parameters = append(parameters, "sensor")
			parameterValue = append(parameterValue, value[0])
		case "start":
			if len(value) != 1 {
				return nil, fmt.Errorf("unexpected length of the value paramter %s has %d attributes", parameter, len(value))
			}

			parsedValue, err := strconv.ParseInt(value[0], 10, 64)
			if err != nil {
				return nil, err
			}

			start = tim.Unix(parsedValue, 0)
		case "end":
			if len(value) != 1 {
				return nil, fmt.Errorf("unexpected length of the value paramter %s has %d attributes", parameter, len(value))
			}
			parsedValue, err := strconv.ParseInt(value[0], 10, 64)
			if err != nil {
				return nil, err
			}

			end = tim.Unix(parsedValue, 0)
		}
	}

	var time []tim.Time
	var machine []string
	var id []int64
	var cTime, cMachine, cId interface{}
	cTime = time
	cMachine = machine
	cId = id
	values := []*interface{}{
		&cTime,
		&cMachine,
		&cId,
	}
	if err := db.QueryTime("analyse_result", []string{"id", "time", "machine"}, parameters, "time", start, end, values, parameterValue); err != nil {
		return nil, err
	}
	time = cTime.([]tim.Time)
	machine = cMachine.([]string)
	id = cId.([]int64)

	if len(time) != len(machine) && len(machine) != len(id) {
		return nil, fmt.Errorf("the length of the array of the database result has not the same size")
	}

	var ret []models.ResultList
	for i := 0; i < len(time); i++ {
		date := time[i]
		ret = append(ret, models.ResultList{
			Id:      id[i],
			Machine: machine[i],
			Date:    date.Unix(),
		})
	}
	return ret, nil
}
