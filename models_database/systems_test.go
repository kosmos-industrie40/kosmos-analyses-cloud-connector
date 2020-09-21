package models_database

import (
	"database/sql/driver"
	"fmt"
	"testing"

	dbMock "github.com/DATA-DOG/go-sqlmock"
)

func TestSystemsExists(t *testing.T) {

	testTable := []struct {
		rows              *dbMock.Rows
		system            System
		args              []interface{}
		expectedExistence bool
		expectedId        int64
		description       string
	}{
		{
			dbMock.NewRows([]string{"id"}),
			System{},
			[]interface{}{""},
			false,
			0,
			"no matched elemend found",
		},
		{
			dbMock.NewRows([]string{"id"}).AddRow(1),
			System{Name: "abc"},
			[]interface{}{"abc"},
			true,
			1,
			"1 matched elemend found",
		},
		{
			dbMock.NewRows([]string{"id"}).AddRow(1).AddRow(5),
			System{Name: "abc"},
			[]interface{}{"abc"},
			true,
			1,
			"2 matched elemend found",
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("can not open mocked database: %s\n", err)
			}
			defer db.Close()
			mock.ExpectQuery("^SELECT id FROM systems").
				WithArgs(v.args[0]).
				WillReturnRows(v.rows)

			exists, id, err := v.system.Exists(db)

			if err != nil {
				t.Errorf("expected Error != returned error\n\tnil != %s\n", err)
			}

			if id != v.expectedId {
				t.Errorf("expectedId != returned id\n\t%d != %d\n", v.expectedId, id)
			}

			if exists != v.expectedExistence {
				t.Errorf("expectedExistence != returned existence\n\t%t != %t\n", v.expectedExistence, exists)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectaions were met: %s\n", err)
			}

		})
	}
}

func TestSystemsInsert(t *testing.T) {

	testTable := []struct {
		returnValue driver.Result
		systems     System
		args        []interface{}
		expectedId  int64
		err         error
		description string
	}{
		{
			dbMock.NewResult(1, 1),
			System{Name: "abc"},
			[]interface{}{"abc"},
			1,
			nil,
			"INSERT first Systems",
		},
		{
			dbMock.NewErrorResult(fmt.Errorf("bla")),
			System{Name: "abc"},
			[]interface{}{"abc"},
			0,
			fmt.Errorf("bla"),
			"INSERT first Systems",
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("can not open mocked database: %s\n", err)
			}
			defer db.Close()
			mock.ExpectExec("^INSERT INTO systems (.)+").WithArgs(v.args[0]).WillReturnResult(v.returnValue)

			id, err := v.systems.Insert(db)
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

func TestSystemsQuery(t *testing.T) {

	testTable := []struct {
		rows        *dbMock.Rows
		systems     System
		description string
		err         error
	}{
		{
			dbMock.NewRows([]string{"name"}).AddRow("abc"),
			System{
				Name: "abc",
			},
			"no matched elemend found",
			nil,
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("can not open mocked database: %s\n", err)
			}
			defer db.Close()

			mock.ExpectQuery("^SELECT name FROM systems").
				WithArgs(1).
				WillReturnRows(v.rows)

			var systems System
			err = (&systems).Query(db, 1)

			if err != nil && v.err != nil {
				if err.Error() != v.err.Error() {
					t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
				}
			} else if err != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
			} else if v.err != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
			}

			if v.systems.Name != systems.Name {
				t.Errorf("the name are different\n\t%s != %s", v.systems.Name, systems.Name)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectaions were met: %s\n", err)
			}
		})
	}
}
