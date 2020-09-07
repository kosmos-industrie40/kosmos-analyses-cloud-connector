package models

import (
	"database/sql/driver"
	"fmt"
	"testing"

	dbMock "github.com/DATA-DOG/go-sqlmock"
)

func testStringArray(a, b []string, t *testing.T) {
	if len(a) != len(b) {
		t.Errorf("length of the string arrays are not equal\n\t%d != %d", len(a), len(b))
	}

	for i := range a {
		if a[i] != b[i] {
			t.Errorf("%d element of the string array are not equal\n\t%s != %s", i, a[i], b[i])
		}
	}
}

func TestArrayToString(t *testing.T) {
	testTable := []struct {
		description string
		data        []string
		expect      string
	}{
		{
			"0 Element",
			[]string{},
			"{}",
		},
		{
			"1 Element",
			[]string{"one"},
			"{\"one\"}",
		},
		{
			"2 Elements",
			[]string{"one", "two"},
			"{\"one\",\"two\"}",
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			output := (&Container{}).arrayToString(v.data)
			if output != v.expect {
				t.Errorf("expected string != returned string\n\t%s != %s\n", v.expect, output)
			}
		})
	}
}

func TestStringToArray(t *testing.T) {
	testTable := []struct {
		description string
		data        string
		expect      []string
	}{
		{
			"0 Element",
			"{}",
			[]string{},
		},
		{
			"1 Element",
			"{\"one\"}",
			[]string{"one"},
		},
		{
			"2 Elements",
			"{\"one\",\"two\"}",
			[]string{"one", "two"},
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			output := (&Container{}).stringToArray(v.data)
			testStringArray(output, v.expect, t)
		})
	}
}

func TestContainerExists(t *testing.T) {

	testTable := []struct {
		rows              *dbMock.Rows
		container         Container
		args              []interface{}
		expectedExistence bool
		expectedId        int64
		description       string
	}{
		{
			dbMock.NewRows([]string{"id"}),
			Container{
				Arguments:   []string{},
				Environment: []string{},
			},
			[]interface{}{"", "", "{}", "{}"},
			false,
			0,
			"no matched elemend found",
		},
		{
			dbMock.NewRows([]string{"id"}).AddRow(1),
			Container{
				Url:         "abc",
				Tag:         "5",
				Arguments:   []string{},
				Environment: []string{},
			},
			[]interface{}{"abc", "5", "{}", "{}"},
			true,
			1,
			"1 matched elemend found",
		},
		{
			dbMock.NewRows([]string{"id"}).AddRow(1).AddRow(5),
			Container{
				Url:         "abc",
				Tag:         "5",
				Arguments:   []string{},
				Environment: []string{},
			},
			[]interface{}{"abc", "5", "{}", "{}"},
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
			mock.ExpectQuery("^SELECT id FROM containers").
				WithArgs(v.args[0], v.args[1], v.args[2], v.args[3]).
				WillReturnRows(v.rows)

			exists, id, err := v.container.Exists(db)

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

func TestContainerInsert(t *testing.T) {

	testTable := []struct {
		returnValue driver.Result
		container   Container
		args        []interface{}
		expectedId  int64
		err         error
		description string
	}{
		{
			dbMock.NewResult(1, 1),
			Container{
				Url:         "abc",
				Tag:         "test",
				Arguments:   []string{},
				Environment: []string{},
			},
			[]interface{}{"abc", "test", "{}", "{}"},
			1,
			nil,
			"INSERT first Container",
		},
		{
			dbMock.NewErrorResult(fmt.Errorf("bla")),
			Container{
				Url:         "abc",
				Tag:         "test",
				Arguments:   []string{},
				Environment: []string{},
			},
			[]interface{}{"abc", "test", "{}", "{}"},
			0,
			fmt.Errorf("bla"),
			"INSERT first Container",
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("can not open mocked database: %s\n", err)
			}
			defer db.Close()
			mock.ExpectExec("^INSERT INTO containers (.)+").WithArgs(v.args[0], v.args[1], v.args[2], v.args[3]).WillReturnResult(v.returnValue)

			id, err := v.container.Insert(db)
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

func TestContainerQuery(t *testing.T) {

	testTable := []struct {
		rows        *dbMock.Rows
		container   Container
		description string
		err         error
	}{
		{
			dbMock.NewRows([]string{"url", "tag", "arguments", "environment"}).AddRow("abc", "cde", "{}", "{}"),
			Container{
				Url:         "abc",
				Tag:         "cde",
				Arguments:   []string{},
				Environment: []string{},
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

			mock.ExpectQuery("^SELECT url, tag, arguments, environment FROM containers").
				WithArgs(1).
				WillReturnRows(v.rows)

			var container Container
			err = (&container).Query(db, 1)

			if err != nil && v.err != nil {
				if err.Error() != v.err.Error() {
					t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
				}
			} else if err != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
			} else if v.err != nil {
				t.Errorf("expected error != returned err\n\t%s != %s", err, v.err)
			}

			if v.container.Url != container.Url {
				t.Errorf("the urls are different\n\t%s != %s", v.container.Url, container.Url)
			}

			if v.container.Tag != container.Tag {
				t.Errorf("the urls are different\n\t%s != %s", v.container.Tag, container.Tag)
			}

			testStringArray(v.container.Arguments, container.Arguments, t)
			testStringArray(v.container.Environment, container.Environment, t)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectaions were met: %s\n", err)
			}
		})
	}
}
