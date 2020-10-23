package auth

import (
	"database/sql/driver"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	dbMock "github.com/DATA-DOG/go-sqlmock"
	"k8s.io/klog"
)

func TestHelperOidc_IsAuthenticated(t *testing.T) {
	testTable := []struct {
		description string
		token       string
		err         error
		statusCode  int
		writeAccess bool
		authSuccess bool
		contract    string
		validRows   *dbMock.Rows
		orgRows     *dbMock.Rows
	}{
		{
			"success no write",
			"token",
			nil,
			0,
			false,
			true,
			"contract",
			dbMock.NewRows([]string{"valid"}).AddRow(time.Now().Add(time.Hour)),
			dbMock.NewRows([]string{"organisations"}).AddRow("org"),
		},
		{
			"success write",
			"token",
			nil,
			0,
			true,
			true,
			"contract",
			dbMock.NewRows([]string{"valid"}).AddRow(time.Now().Add(time.Hour)),
			dbMock.NewRows([]string{"organisations"}).AddRow("org"),
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {

			req, err := http.NewRequest("GET", "url", nil)
			if err != nil {
				klog.Errorf("cannot create http request: %s", err)
			}
			req.Header.Add("token", v.token)

			db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("cannot create dbmock: %s", err)
			}

			mock.ExpectQuery("SELECT valid FROM token WHERE token = $1").
				WithArgs(v.token).
				WillReturnRows(v.validRows)

			var table string
			if v.writeAccess {
				table = "write_permissions rp"
			} else {
				table = "read_permissions rp"
			}

			mock.ExpectQuery(fmt.Sprintf("SELECT tp.organisation FROM token_permission as tp JOIN %s on tp.organisation = rp.organisation WHERE token = $1 AND contract = $2", table)).
				WithArgs(v.token, v.contract).
				WillReturnRows(v.orgRows)

			defer db.Close()

			helper := NewAuthHelper(db, "")
			isAuth, statusCode, err := helper.IsAuthenticated(req, v.contract, v.writeAccess)

			if statusCode != v.statusCode {
				t.Errorf("expected status code != returned status code\n\t%d != %d", v.statusCode, statusCode)
			}

			if isAuth != v.authSuccess {
				t.Errorf("expected auth != returned auth\n\t%t != %t", v.authSuccess, isAuth)
			}

			if !reflect.DeepEqual(err, v.err) {
				t.Errorf("expected error != returned error\n\t%s != %s", v.err, err)
			}
		})
	}
}

func TestHelperOidc_cleanUp(t *testing.T) {
	testTable := []struct {
		description string
		result      driver.Result
		err         error
	}{
		{
			"success",
			dbMock.NewResult(0, 5),
			nil,
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("cannot create dbmock: %s", err)
			}

			defer db.Close()

			mock.ExpectExec("DELETE FROM token").
				WillReturnResult(v.result)

			err = helperOidc{db: db}.cleanUp()

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectations were met: %s", err)
			}

			if !reflect.DeepEqual(err, v.err) {
				t.Errorf("expected error != returned error\n\t%s != %s", err, v.err)
			}
		})
	}
}

func TestHelperOidc_CreateSession(t *testing.T) {
	testTable := []struct {
		description string
		orgs        []struct {
			name string
			id   int64
		}
		orgRows     *dbMock.Rows
		token       string
		valid       time.Time
		tokenResult driver.Result
		err         error
	}{
		{
			"success",
			[]struct {
				name string
				id   int64
			}{
				{
					name: "name",
					id:   4,
				},
			},
			dbMock.NewRows([]string{"id"}).AddRow(4),
			"token",
			time.Now(),
			dbMock.NewResult(0, 1),
			nil,
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("cannot create mock: %s", err)
			}

			var namesOrgs []string
			for _, org := range v.orgs {
				namesOrgs = append(namesOrgs, org.name)
			}

			mock.ExpectQuery("SELECT id FROM organisations WHERE name in ($1)").
				WithArgs(fmt.Sprintf("'%s'", strings.Join(namesOrgs, "','"))).
				WillReturnRows(v.orgRows)

			mock.ExpectExec("INSERT INTO token (token, valid, write_contract) VALUES ($1, $2, $3)").
				WithArgs(v.token, v.valid, false).
				WillReturnResult(v.tokenResult)

			for _, org := range v.orgs {
				mock.ExpectExec("INSERT INTO token_permission (token, organisation) VALUES ($1, $2)").
					WithArgs(v.token, org.id).
					WillReturnResult(dbMock.NewResult(0, 1))
			}

			helper := NewAuthHelper(db, "")
			err = helper.CreateSession(v.token, namesOrgs, []string{}, v.valid)

			if !reflect.DeepEqual(err, v.err) {
				t.Errorf("expected error != returned error\n\t%s != %s", v.err, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectations were met: %s", err)
			}
		})
	}
}

func TestHelperOidc_DeleteSession(t *testing.T) {
	testTable := []struct {
		description string
		result      driver.Result
		token       string
		err         error
	}{
		{
			"delete session success",
			dbMock.NewResult(0, 0),
			"token",
			nil,
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			db, mock, err := dbMock.New()
			if err != nil {
				t.Fatalf("cannot craete dbmock: %s", err)
			}

			defer db.Close()

			mock.ExpectExec("DELETE FROM token").
				WithArgs(v.token).
				WillReturnResult(v.result)

			helper := NewAuthHelper(db, "")
			err = helper.DeleteSession(v.token)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectaions were met: %s", err)
			}

			if !reflect.DeepEqual(err, v.err) {
				t.Errorf("expected error != returned err\n\t%s != %s", v.err, err)
			}
		})
	}
}

func TestHelperOidc_TokenValid(t *testing.T) {
	testTable := []struct {
		description string
		token       string
		err         error
		tokenValid  bool
		rows        *dbMock.Rows
	}{
		{
			"empty token",
			"",
			nil,
			false,
			nil,
		},
		{
			"success",
			"token",
			nil,
			true,
			dbMock.NewRows([]string{"token", "valid"}).AddRow("token", "2020-08-21T11:00:88.123Z"),
		},
		{
			"nothing found",
			"token",
			nil,
			false,
			dbMock.NewRows([]string{"token", "valid"}),
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "something", nil)
			if err != nil {
				t.Fatalf("cannot create http request")
			}

			req.Header.Set("token", v.token)

			db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("cannot create mocked databse: %s", err)
			}

			if v.token != "" {
				mock.ExpectQuery("SELECT * FROM token WHERE token = $1 AND valid >= NOW()").
					WithArgs(v.token).
					WillReturnRows(v.rows)
			}

			defer db.Close()

			helper := helperOidc{db: db}
			valid, err := helper.TokenValid(req)

			if !reflect.DeepEqual(err, v.err) {
				klog.Errorf("returned error != expected error\n\t%s != %s", err, v.err)
			}

			if valid != v.tokenValid {
				klog.Errorf("returned validation != expected validation\n\t%t != %t", valid, v.tokenValid)
			}
		})
	}
}
