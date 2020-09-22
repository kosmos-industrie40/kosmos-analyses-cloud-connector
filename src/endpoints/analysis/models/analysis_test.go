package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAnalysis_Validate(t *testing.T) {
	rootDir := "../../../../kosmos-json-specifications/mqtt_payloads/"
	err := filepath.Walk(rootDir,
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				t.Fatalf("cannot walk through the files: %s", err)
			}

			match, err := filepath.Match(rootDir+"analysis-example-*.json", path)
			if err != nil {
				t.Fatalf("cannot execute filepath matcher: %s", err)
			}

			if match {
				t.Run(fmt.Sprintf("validating with file: %s", path), func(t *testing.T) {
					data, err := ioutil.ReadFile(path)
					if err != nil {
						t.Fatalf("cannot read file: %s", err)
					}

					var ana Analysis

					err = json.Unmarshal(data, &ana)
					if err != nil {
						t.Fatalf("cannot unmarshal file %s: %s ", path, err)
					}

					if !ana.Validate() {
						t.Errorf("cannot validate %s", path)
					}
				})
			}

			return nil
		},
	)

	if err != nil {
		t.Errorf("filepath.Walk returned error: %s", err)
	}
}

var ana = Analysis{
	From:      "from",
	Timestamp: "timestamp",
	Model: Model{
		URL: "url",
		Tag: "tag",
	},
	Type: "text",
	Calculated: struct {
		Message struct {
			Machine string `json:"machine"`
			Sensor  string `json:"sensor"`
		} `json:"message"`
		Received string `json:"received"`
	}{
		Message: struct {
			Machine string `json:"machine"`
			Sensor  string `json:"sensor"`
		}{
			Machine: "machine",
			Sensor:  "sensor",
		},
		Received: "",
	},
	Results:   nil,
	Signature: "",
}

func TestAnalysisHandler_Insert(t *testing.T) {
	testTable := []struct {
		description string
		contractID  string
		machineID   string
		sensorID    string
		cmsID       int64
		cmsIdSQL    *sqlmock.Rows
		err         error
		analysis    Analysis
		result      driver.Result
	}{
		{
			"successfully insertion",
			"contract",
			"machine",
			"sensor",
			6,
			sqlmock.NewRows([]string{"id"}).AddRow(6),
			nil,
			ana,
			sqlmock.NewResult(0, 0),
		},
		{
			"error not matching contract-machine-sensor",
			"contract",
			"machine",
			"sensor",
			0,
			sqlmock.NewRows([]string{"id"}),
			fmt.Errorf("no matching contract-machine-sensor combination found"),
			Analysis{},
			sqlmock.NewResult(0, 0),
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("cannot create mocked databse: %s", err)
			}

			defer db.Close()

			data, err := json.Marshal(v.analysis)
			if err != nil {
				t.Fatalf("cannot marshal analysis")
			}

			mock.ExpectQuery("SELECT cms.id FROM contract_machine_sensors AS cms JOIN machine_sensors ms on cms.machine_sensor = ms.id JOIN sensors s on ms.sensor = s.id WHERE cms.contract = $1 AND ms.machine = $2 AND s.transmitted_id = $3").
				WithArgs(v.contractID, v.machineID, v.sensorID).
				WillReturnRows(v.cmsIdSQL)

			mock.ExpectExec("INSERT INTO analyse_result (contract_machine_sensor, time, result) VALUES ($1, $2, $3)").
				WithArgs(v.cmsID, v.analysis.Timestamp, string(data)).
				WillReturnResult(v.result)

			aH := analysisHandler{db: db}
			err = aH.Insert(v.contractID, v.machineID, v.sensorID, v.analysis)
			if err != nil && v.err != nil {
				if err.Error() != v.err.Error() {
					t.Errorf("returned error != expected error\n\t%s != %s", err, v.err)
				}
			} else if err != nil || v.err != nil {
				t.Errorf("returned error != expected error\n\t%s != %s", err, v.err)
			}
		})
	}
}

func TestAnalysisHandler_Query(t *testing.T) {
	anaJson, err := json.Marshal(ana)
	if err != nil {
		t.Fatalf("cannot convert analysis object to json: %s", err)
	}
	testTable := []struct {
		description string
		contractId  string
		resultId    int64
		rows        *sqlmock.Rows
		err         error
		analysis    Analysis
	}{
		{
			"successfully query",
			"contract",
			1,
			sqlmock.NewRows([]string{"result"}).AddRow(anaJson),
			nil,
			ana,
		},
		{
			"successfully not found query",
			"contract",
			1,
			sqlmock.NewRows([]string{"result"}),
			nil,
			Analysis{},
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Errorf("cannot create mocked database: %s", err)
			}

			defer db.Close()

			mock.ExpectQuery("SELECT result FROM analyse_result AS ar JOIN contract_machine_sensors cms on ar.contract_machine_sensor = cms.id").
				WithArgs(v.contractId, v.resultId).
				WillReturnRows(v.rows)

			aH := analysisHandler{db: db}

			analysis, err := aH.Query(v.contractId, v.resultId)
			if err != nil && v.err != nil {
				if err.Error() != v.err.Error() {
					t.Errorf("returned error != expected error\n\t%s != %s", err, v.err)
				}
			} else if err != nil || v.err != nil {
				t.Errorf("returned error != expected error\n\t%s != %s", err, v.err)
			}

			if !reflect.DeepEqual(analysis, v.analysis) {
				t.Errorf("returned analysis and expected analysis are not equal")
			}
		})
	}
}
