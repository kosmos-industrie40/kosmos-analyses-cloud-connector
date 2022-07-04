package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	dbMock "github.com/DATA-DOG/go-sqlmock"
)

func TestDeleteContract(t *testing.T) {
	testTable := []struct {
		description string
		contractID  string
		expectError error
	}{
		{
			"success",
			"contract",
			nil,
		},
		{
			"db query error",
			"contract",
			fmt.Errorf("error"),
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("cannot create database mock: %s", err)
			}

			defer db.Close()

			if v.expectError == nil {
				mock.ExpectExec("UPDATE contracts SET active = false WHERE id = $1").WithArgs(v.contractID).WillReturnResult(dbMock.NewResult(0, 0))
			} else {
				mock.ExpectExec("UPDATE contracts SET active = false WHERE id = $1").WithArgs(v.contractID).WillReturnError(v.expectError)
			}

			handler := contractHandler{db: db}
			err = handler.DeleteContract(v.contractID)

			if v.expectError != nil && err != nil {
				if v.expectError.Error() != err.Error() {
					t.Errorf("returned error is not equal to expected error\n\t%s != %s", v.expectError, err)
				}
			} else if v.expectError != nil || err != nil {
				t.Errorf("returned error is not equal to expected error\n\t%s != %s", v.expectError, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectaions are met: %s", err)
			}
		})
	}
}

func TestGetContract(t *testing.T) {
	testSuccessContract := Contract{}
	testSuccessContract.Body.Machine = "abc"
	jsonTestSuccessContract, err := json.Marshal(testSuccessContract)
	if err != nil {
		t.Fatalf("cannot marshal contract: %s", err)
	}

	testTable := []struct {
		description   string
		contractID    string
		contract      Contract
		dbError       error
		expectedError error
		rows          *dbMock.Rows
	}{
		{
			"success",
			"contract",
			testSuccessContract,
			nil,
			nil,
			dbMock.NewRows([]string{"contract"}).AddRow(string(jsonTestSuccessContract)),
		},
		{
			"empty",
			"contract",
			Contract{},
			nil,
			nil,
			dbMock.NewRows([]string{"contract"}),
		},
		{
			"empty",
			"contract",
			Contract{},
			fmt.Errorf("error"),
			fmt.Errorf("error"),
			nil,
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("cannot create db mock: %s", err)
			}

			defer db.Close()

			if v.dbError == nil {
				mock.ExpectQuery("SELECT contract FROM contracts").WillReturnRows(v.rows)
			} else {
				mock.ExpectQuery("SELECT contract FROM contracts").WillReturnError(v.dbError)
			}

			handler := contractHandler{db: db}
			contract, err := handler.GetContract(v.contractID)

			if err != nil && v.dbError != nil {
				if err.Error() != v.expectedError.Error() {
					t.Errorf("returned error is not equal to expected error\n\t%s != %s", v.expectedError, err)
				}
			} else if err != nil || v.dbError != nil {
				t.Errorf("returned error is not equal to expected error\n\t%s != %s", v.expectedError, err)
			}

			if !reflect.DeepEqual(contract, v.contract) {
				t.Errorf("returned contract != expected contract\n\t%v != %v", contract, v.contract)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectations were met: %s", err)
			}

		})
	}
}

func TestInsertStorageDuration(t *testing.T) {
	testTable := []struct {
		description           string
		contractMachineSensor int64
		duration              string
		system                int64
		expectedError         error
	}{
		{
			"success",
			3,
			"2020-10-09T11:34:22Z",
			2,
			nil,
		},
		{
			"err",
			3,
			"2020-10-09T11:34:22Z",
			2,
			fmt.Errorf("error"),
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("cannot create mocked database: %s", err)
			}

			defer db.Close()

			if v.expectedError == nil {
				mock.ExpectExec("INSERT INTO storage_duration (system, contract_machine_sensor, duration) VALUES ($1, $2, $3)").WithArgs(v.system, v.contractMachineSensor, v.duration).WillReturnResult(dbMock.NewResult(1, 1))
			} else {
				mock.ExpectExec("INSERT INTO storage_duration (system, contract_machine_sensor, duration) VALUES ($1, $2, $3)").WithArgs(v.system, v.contractMachineSensor, v.duration).WillReturnError(v.expectedError)
			}

			handler := contractHandler{db: db}
			err = handler.insertStorageDuration(v.contractMachineSensor, v.duration, v.system)
			if err != nil && v.expectedError != nil {
				if err.Error() != v.expectedError.Error() {
					t.Errorf("returned error != expected Error\n\t%s != %s", err, v.expectedError)
				}
			} else if err != nil || v.expectedError != nil {
				t.Errorf("returned error != expected Error\n\t%s != %s", err, v.expectedError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expected sql queries were matched")
			}
		})
	}
}
