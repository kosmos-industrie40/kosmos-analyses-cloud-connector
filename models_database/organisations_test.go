package models_database

import (
	"database/sql/driver"
	"fmt"
	"testing"

	dbMock "github.com/DATA-DOG/go-sqlmock"
)

func TestOrganistaionExists(t *testing.T) {

	testTable := []struct {
		rows              *dbMock.Rows
		organisation      Organisation
		args              []interface{}
		expectedExistence bool
		expectedId        int64
		description       string
	}{
		{
			dbMock.NewRows([]string{"id"}),
			Organisation{},
			[]interface{}{""},
			false,
			0,
			"no matched elemend found",
		},
		{
			dbMock.NewRows([]string{"id"}).AddRow(1),
			Organisation{Name: "abc"},
			[]interface{}{"abc"},
			true,
			1,
			"1 matched elemend found",
		},
		{
			dbMock.NewRows([]string{"id"}).AddRow(1).AddRow(5),
			Organisation{Name: "abc"},
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
			mock.ExpectQuery("^SELECT id FROM organisations").
				WithArgs(v.args[0]).
				WillReturnRows(v.rows)

			exists, id, err := v.organisation.Exists(db)

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

func TestOrganistaionInsert(t *testing.T) {

	testTable := []struct {
		returnValue   driver.Result
		organisations Organisation
		args          []interface{}
		expectedId    int64
		err           error
		description   string
	}{
		{
			dbMock.NewResult(1, 1),
			Organisation{Name: "abc"},
			[]interface{}{"abc"},
			1,
			nil,
			"INSERT first Organisations",
		},
		{
			dbMock.NewErrorResult(fmt.Errorf("bla")),
			Organisation{Name: "abc"},
			[]interface{}{"abc"},
			0,
			fmt.Errorf("bla"),
			"INSERT first Organisations",
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("can not open mocked database: %s\n", err)
			}
			defer db.Close()
			mock.ExpectExec("^INSERT INTO organisations (.)+").WithArgs(v.args[0]).WillReturnResult(v.returnValue)

			id, err := v.organisations.Insert(db)
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

func TestOrganisationsQuery(t *testing.T) {

	testTable := []struct {
		rows          *dbMock.Rows
		organisations Organisation
		description   string
		err           error
	}{
		{
			dbMock.NewRows([]string{"name"}).AddRow("abc"),
			Organisation{
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

			mock.ExpectQuery("^SELECT name FROM organisations").
				WithArgs(1).
				WillReturnRows(v.rows)

			var organisations Organisation
			err = (&organisations).Query(db, 1)

			if err != nil && v.err != nil {
				if err.Error() != v.err.Error() {
					t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
				}
			} else if err != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
			} else if v.err != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
			}

			if v.organisations.Name != organisations.Name {
				t.Errorf("the name are different\n\t%s != %s", v.organisations.Name, organisations.Name)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectaions were met: %s\n", err)
			}
		})
	}
}
