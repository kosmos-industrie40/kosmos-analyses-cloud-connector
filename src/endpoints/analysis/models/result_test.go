package models

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestResultListHandler_Get(t *testing.T) {
	testTable := []struct {
		description string
		contractID  string
		params      map[string][]string
		err         error
		ret         []byte
		args        []driver.Value
		rows        *sqlmock.Rows
		machineRow  *sqlmock.Rows
		sensorRow   *sqlmock.Rows
	}{
		{
			"empty get without params",
			"contract",
			nil,
			nil,
			[]byte("null"),
			[]driver.Value{"contract"},
			sqlmock.NewRows([]string{"id", "time", "machine"}),
			nil,
			nil,
		},
		{
			"get without params",
			"contract",
			nil,
			nil,
			[]byte("[{\"resultID\":4,\"machine\":\"mach\",\"date\":\"data\"}]"),
			[]driver.Value{"contract"},
			sqlmock.NewRows([]string{"id", "time", "machine"}).AddRow(4, "data", "mach"),
			nil,
			nil,
		},
		{
			"get two results without params",
			"contract",
			nil,
			nil,
			[]byte("[{\"resultID\":4,\"machine\":\"mach\",\"date\":\"data\"},{\"resultID\":5,\"machine\":\"mach\",\"date\":\"data\"}]"),
			[]driver.Value{"contract"},
			sqlmock.NewRows([]string{"id", "time", "machine"}).AddRow(4, "data", "mach").AddRow(5, "data", "mach"),
			nil,
			nil,
		},
		{
			"empty get with not nil params",
			"contract",
			map[string][]string{},
			nil,
			[]byte("null"),
			[]driver.Value{"contract"},
			sqlmock.NewRows([]string{"id", "time", "machine"}),
			nil,
			nil,
		},
		{
			"empty get with start array != 1",
			"contract",
			map[string][]string{
				"start": {"2020-09-23T10:24:55Z", "b"},
			},
			fmt.Errorf("unexpected length of the query parameters"),
			[]byte{},
			[]driver.Value{"contract", "2020-09-23T10:24:55Z"},
			sqlmock.NewRows([]string{"id", "time", "machine"}),
			nil,
			nil,
		},
		{
			"empty get with start",
			"contract",
			map[string][]string{
				"start": {"2020-09-23T10:24:55Z"},
			},
			nil,
			[]byte{},
			[]driver.Value{"contract", "2020-09-23T10:24:55Z"},
			sqlmock.NewRows([]string{"id", "time", "machine"}),
			nil,
			nil,
		},
		{
			"get with start",
			"contract",
			map[string][]string{
				"start": {"2020-09-23T10:24:55Z"},
			},
			nil,
			[]byte("[{\"resultID\":4,\"machine\":\"mach\",\"date\":\"data\"}]"),
			[]driver.Value{"contract", "2020-09-23T10:24:55Z"},
			sqlmock.NewRows([]string{"id", "time", "machine"}).AddRow(4, "data", "mach"),
			nil,
			nil,
		},
		{
			"get with start 2 results",
			"contract",
			map[string][]string{
				"start": {"2020-09-23T10:24:55Z"},
			},
			nil,
			[]byte("[{\"resultID\":4,\"machine\":\"mach\",\"date\":\"data\"},{\"resultID\":5,\"machine\":\"mach\",\"date\":\"data\"}]"),
			[]driver.Value{"contract", "2020-09-23T10:24:55Z"},
			sqlmock.NewRows([]string{"id", "time", "machine"}).AddRow(4, "data", "mach").AddRow(5, "data", "mach"),
			nil,
			nil,
		},
		{
			"get with end",
			"contract",
			map[string][]string{
				"end": {"2020-09-23T10:24:55Z"},
			},
			nil,
			[]byte("[{\"resultID\":4,\"machine\":\"mach\",\"date\":\"data\"}]"),
			[]driver.Value{"contract", "2020-09-23T10:24:55Z"},
			sqlmock.NewRows([]string{"id", "time", "machine"}).AddRow(4, "data", "mach"),
			nil,
			nil,
		},
		{
			"get with machine",
			"contract",
			map[string][]string{
				"machine": {"machine"},
			},
			nil,
			[]byte("[{\"resultID\":4,\"machine\":\"mach\",\"date\":\"data\"}]"),
			[]driver.Value{"contract", "5"},
			sqlmock.NewRows([]string{"id", "time", "machine"}).AddRow(4, "data", "mach"),
			sqlmock.NewRows([]string{"machine"}).AddRow(5),
			nil,
		},
		{
			"get with sensor",
			"contract",
			map[string][]string{
				"sensor": {"sensor"},
			},
			nil,
			[]byte("[{\"resultID\":4,\"machine\":\"mach\",\"date\":\"data\"}]"),
			[]driver.Value{"contract", "5"},
			sqlmock.NewRows([]string{"id", "time", "machine"}).AddRow(4, "data", "mach"),
			nil,
			sqlmock.NewRows([]string{"sensor"}).AddRow(5),
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("cannot create mocked database: %s", err)
			}

			defer db.Close()

			if v.machineRow != nil {
				mock.ExpectQuery("SELECT cms.id FROM contract_machine_sensors as cms JOIN machine_sensors ms on cms.machine_sensor = ms.id").
					WithArgs("machine").
					WillReturnRows(v.machineRow)
			}

			if v.sensorRow != nil {
				mock.ExpectQuery("SELECT cms.id FROM contract_machine_sensors as cms JOIN machine_sensors ms on cms.machine_sensor = ms.id JOIN sensors s on ms.sensor = s.id").
					WithArgs("sensor").
					WillReturnRows(v.sensorRow)
			}

			mock.ExpectQuery("SELECT ar.id, ar.time, ms.machine FROM analysis_result AS ar JOIN contract_machine_sensors cms on cms.id = ar.contract_machine_sensor JOIN machine_sensors ms on cms.machine_sensor = ms.id").
				WithArgs(v.args...).
				WillReturnRows(v.rows)

			data, err := NewResultList(db).Get(v.contractID, v.params)

			if len(data) != len(v.ret) && len(v.ret) != 0 {
				if !reflect.DeepEqual(data, v.ret) {
					t.Errorf("returned data != expected data\n\t%s != %s", string(data), string(v.ret))
				}
			}

			if !reflect.DeepEqual(err, v.err) {
				t.Errorf("expected error != returned error\n\t%s != %s", v.err, err)
			}
		})
	}
}
