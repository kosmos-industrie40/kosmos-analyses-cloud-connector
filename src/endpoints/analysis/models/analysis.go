package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/klog"
)

type AnalysisHandler interface {
	Insert(string, string, string, Analysis) error
	Query(string, int64) (Analysis, error)
}

type analysisHandler struct {
	db *sql.DB
}

func NewAnalysisHandler(db *sql.DB) AnalysisHandler {
	return analysisHandler{db: db}
}

func (a analysisHandler) Insert(contractID string, machineID string, sensorID string, analysis Analysis) error {
	query, err := a.db.Query("SELECT cms.id FROM contract_machine_sensors AS cms JOIN machine_sensors ms on cms.machine_sensor = ms.id JOIN sensors s on ms.sensor = s.id WHERE cms.contract = $1 AND ms.machine = $2 AND s.transmitted_id = $3",
		contractID,
		machineID,
		sensorID,
	)

	if err != nil {
		return err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	var cmsId int64
	if !query.Next() {
		return fmt.Errorf("no matching contract-machine-sensor combination found")
	}

	if err := query.Scan(&cmsId); err != nil {
		return err
	}

	data, err := json.Marshal(analysis)
	if err != nil {
		return err
	}

	_, err = a.db.Exec("INSERT INTO analyse_result (contract_machine_sensor, time, result) VALUES ($1, $2, $3)", cmsId, analysis.Timestamp, string(data))
	return err
}

func (a analysisHandler) Query(contractID string, resultID int64) (Analysis, error) {
	query, err := a.db.Query("SELECT result FROM analyse_result AS ar JOIN contract_machine_sensors cms on ar.contract_machine_sensor = cms.id WHERE ar.id = $1 AND cms.contract = $2",
		contractID,
		resultID,
	)
	if err != nil {
		return Analysis{}, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s", err)
		}
	}()

	if !query.Next() {
		return Analysis{}, nil
	}

	var data string
	if err := query.Scan(&data); err != nil {
		return Analysis{}, err
	}

	var analysis Analysis
	err = json.Unmarshal([]byte(data), &analysis)
	return analysis, err
}

type Analysis struct {
	From       string `json:"from"`
	Timestamp  string `json:"timestamp"`
	Model      Model  `json:"model"`
	Type       string `json:"type"`
	Calculated struct {
		Message struct {
			Machine string `json:"machine"`
			Sensor  string `json:"sensor"`
		} `json:"message"`
		Received string `json:"received"`
	} `json:"calculated"`
	Results   interface{} `json:"results"`
	Signature string      `json:"signature"`
}

func (a Analysis) Validate() bool {

	if _, err := time.Parse(time.RFC3339, a.Timestamp); err != nil {
		return false
	}
	result, err := json.Marshal(a.Results)
	if err != nil {
		klog.Errorf("unexpected error in marshaling analysis: %s", err)
		return false
	}

	switch a.Type {
	case "text":
		var text TextResult
		if err := json.Unmarshal(result, &text); err != nil {
			return false
		}
	case "time_series":
		var timeSeriesResult TimeSeriesResult
		if err := json.Unmarshal(result, &timeSeriesResult); err != nil {
			return false
		}
	case "multiple_time_series":
		var mTimeSeriesResult []TimeSeriesResult
		if err := json.Unmarshal(result, &mTimeSeriesResult); err != nil {
			return false
		}
	default:
		return false
	}

	return true
}
