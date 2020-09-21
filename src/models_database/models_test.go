package models_database

import (
	"database/sql/driver"
	"fmt"
	"testing"

	dbMock "github.com/DATA-DOG/go-sqlmock"
)

func TestModel_Insert(t *testing.T) {
	testTable := []struct {
		modelsResult    driver.Result
		containerRows   *dbMock.Rows
		containerResult driver.Result // using listInsertID == 0 to identify, that this is not used in this test
		container       Container
		argsModels      []interface{}
		expectedId      int64
		err             error
		description     string
	}{
		{
			dbMock.NewResult(1, 1),
			dbMock.NewRows([]string{"id"}),
			dbMock.NewResult(1, 1),
			Container{Tag: "tag", Url: "url"},
			[]interface{}{1},
			1,
			nil,
			"test insert both models_database and containers",
		},
		{
			dbMock.NewResult(5, 1),
			dbMock.NewRows([]string{"id"}).AddRow(2),
			dbMock.NewResult(0, 0),
			Container{Tag: "tag", Url: "url"},
			[]interface{}{2},
			5,
			nil,
			"test insert models_database only",
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("can not open mocked database %s\n", err)
			}

			defer db.Close()

			mock.ExpectQuery("^SELECT id FROM containers").WillReturnRows(v.containerRows)
			if id, _ := v.containerResult.LastInsertId(); id != 0 {
				mock.ExpectExec("INSERT INTO containers (.)+").WillReturnResult(v.containerResult)
			}
			mock.ExpectExec("INSERT INTO models_database (.)+").WillReturnResult(v.modelsResult)

			model := Model{Container: v.container}
			id, err := model.Insert(db)

			if err != nil && v.err != nil {
				if err.Error() != v.err.Error() {
					t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
				}
			} else if err != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
			} else if v.err != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
			}

			if id != v.expectedId {
				t.Errorf("expectedId != returned id\n\t%d != %d\n", v.expectedId, id)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectaions were met: %s\n", err)
			}
		})
	}
}

func TestModel_Query(t *testing.T) {
	testTable := []struct {
		modelRows         *dbMock.Rows
		containerRows     *dbMock.Rows
		bothQuery         bool
		expectedContainer Container
		expectedError     error
		description       string
	}{
		{
			dbMock.NewRows([]string{"container"}),
			dbMock.NewRows([]string{"url", "tag", "arguments", "environment"}),
			false,
			Container{},
			fmt.Errorf("no matching model found"),
			"expect error no matching model found",
		},
		{
			dbMock.NewRows([]string{"container"}).AddRow(4),
			dbMock.NewRows([]string{"url", "tag", "arguments", "environment"}),
			true,
			Container{},
			fmt.Errorf("no container found to id: 4"),
			"expect error no matching model found",
		},
		{
			dbMock.NewRows([]string{"container"}).AddRow(4),
			dbMock.NewRows([]string{"url", "tag", "arguments", "environment"}).AddRow("url", "tag", "{}", "{}"),
			true,
			Container{Url: "url", Tag:"tag"},
			nil,
			"successful creation",
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T){
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("can not open mocked database %s\n", err)
			}

			defer db.Close()

			mock.ExpectQuery("SELECT container FROM models_database").WithArgs(1).WillReturnRows(v.modelRows)
			if v.bothQuery {
				mock.ExpectQuery("SELECT url, tag, arguments, environment FROM containers").WillReturnRows(v.containerRows)
			}

			var model Model
			err = model.Query(db, 1)

			if err != nil && v.expectedError != nil {
				if err.Error() != v.expectedError.Error() {
					t.Errorf("expected error != returned err\n\t%s != %s", err, v.expectedError)
				}
			} else if err != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.expectedError)
			} else if v.expectedError != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.expectedError)
			}


			if model.Container.Url != v.expectedContainer.Url || model.Container.Tag != v.expectedContainer.Tag {
				t.Errorf("expected container and return container are not equal")
			}

			testStringArray(model.Container.Arguments, v.expectedContainer.Arguments, t)
			testStringArray(model.Container.Environment, v.expectedContainer.Environment, t)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectaions were met: %s\n", err)
			}
		})
	}
}

func TestModel_Exists(t *testing.T) {
	testTable := []struct{
		modelRows *dbMock.Rows
		containerRows *dbMock.Rows
		bothQuery bool
		expectedID int64
		expectedError error
		expectedExistence bool
		description string
	}{
		{
			dbMock.NewRows([]string{"id"}),
			dbMock.NewRows([]string{"url", "tag", "arguments", "environment"}),
			false,
			0,
			nil,
			false,
			"container doesn't exists",
		},
		{
			dbMock.NewRows([]string{"id"}),
			dbMock.NewRows([]string{"id"}).AddRow(1),
			true,
			0,
			nil,
			false,
			"model doesn't exists",
		},
		{
			dbMock.NewRows([]string{"id"}).AddRow(1),
			dbMock.NewRows([]string{"id"}).AddRow(1),
			true,
			1,
			nil,
			true,
			"all exists",
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("can not open mocked database %s\n", err)
			}

			defer db.Close()

			mock.ExpectQuery("SELECT id FROM containers").WillReturnRows(v.containerRows)
			if v.bothQuery {
				mock.ExpectQuery("SELECT id FROM models_database").WithArgs(1).WillReturnRows(v.modelRows)
			}

			var model Model
			exists, id, err := model.Exists(db)

			if err != nil && v.expectedError != nil {
				if err.Error() != v.expectedError.Error() {
					t.Errorf("expected error != returned err\n\t%s != %s", err, v.expectedError)
				}
			} else if err != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.expectedError)
			} else if v.expectedError != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.expectedError)
			}

			if id != v.expectedID {
				t.Errorf("expectedId != returned id\n\t%d != %d\n", v.expectedID, id)
			}

			if exists != v.expectedExistence {
				t.Errorf("expectedExistence != returned existence\n\t%t != %t\n", v.expectedExistence, exists)
			}
		})
	}
}