// build +unit
package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/models_database"
)

type testModel struct{}

func (m testModel) Model(_ database.Postgres) {}

func (m testModel) GetModel(contract string) ([]models_database.Model, error) {
	switch contract {
	default:
		return nil, nil
	case "error":
		return nil, fmt.Errorf("error")
	case "empty":
		return []models_database.Model{}, nil
	case "two":
		ret := []models_database.Model{
			{
				Tag: "tag",
				Url: "url",
			},
			{
				Tag: "tag2",
				Url: "url2",
			},
		}
		return ret, nil
	case "one":
		ret := []models_database.Model{
			{
				Tag: "tag",
				Url: "url",
			},
		}
		return ret, nil
	}
}

func (m testModel) UpdateModel(contract string, model models_database.UpdateModelState) error {
	if contract == "error" {
		return fmt.Errorf("error")
	}
	return nil
}

var model Model = Model{Auth: AuthTest{}, Model: testModel{}}

func TestModelUpdate(t *testing.T) {
	mod, err := json.Marshal(models_database.UpdateModelState{})
	if err != nil {
		t.Fatal(t)
	}
	//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/4
	tstCases := []struct {
		StatusCode int
		Data       string
		Path       string
	}{
		{
			StatusCode: 400,
			Data:       "",
			Path:       "/",
		},
		{
			StatusCode: 500,
			Data:       string(mod),
			Path:       "/contract/error",
		},
		{
			StatusCode: 400,
			Data:       "",
			Path:       "/contract/er",
		},
		{
			StatusCode: 201,
			Data:       string(mod),
			Path:       "/contract/er",
		},
	}

	t.Run("test model update", func(t *testing.T) {
		for _, test := range tstCases {
			req, err := http.NewRequest("PUT", test.Path, strings.NewReader(test.Data))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("token", "abc")

			rr := httptest.NewRecorder()
			model.ServeHTTP(rr, req)

			if status := rr.Code; status != test.StatusCode {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, test.StatusCode)
			}

		}
	})
}

func TestModelGet(t *testing.T) {
	//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/4
	testCases := []struct {
		StatusCode int
		Expected   []models_database.Model
		Path       string
	}{
		{
			StatusCode: 400,
			Path:       "/",
			Expected:   []models_database.Model{},
		},
		{
			StatusCode: 500,
			Path:       "/contract/error",
			Expected:   []models_database.Model{},
		},
		{
			StatusCode: 200,
			Path:       "/contract/empty",
			Expected:   []models_database.Model{},
		},
		{
			StatusCode: 200,
			Path:       "/contract/two",
			Expected: []models_database.Model{
				{
					Tag: "tag",
					Url: "url",
				},
				{
					Tag: "tag2",
					Url: "url2",
				},
			},
		},
		{
			StatusCode: 200,
			Path:       "/contract/one",
			Expected: []models_database.Model{
				{
					Tag: "tag",
					Url: "url",
				},
			},
		},
	}

	t.Run("test model receive", func(t *testing.T) {
		for _, test := range testCases {
			req, err := http.NewRequest("GET", test.Path, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("token", "abc")

			rr := httptest.NewRecorder()
			model.ServeHTTP(rr, req)

			if status := rr.Code; status != test.StatusCode {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, test.StatusCode)
			}

			if len(test.Expected) == 0 {
				if "" != rr.Body.String() {
					t.Errorf("handler returnes wrong body response: got %s want %s", rr.Body.String(), "")
				}
				continue
			}

			ex, err := json.Marshal(test.Expected)
			if err != nil {
				t.Fatal(err)
			}

			if string(ex) != rr.Body.String() {
				t.Errorf("handler returnes wrong body response: got %s want %s", rr.Body.String(), string(ex))
			}
		}
	})
}

func TestModelAuth(t *testing.T) {
	//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/4
	testCases := []struct {
		Token      string
		StatusCode int
	}{
		{
			Token:      "",
			StatusCode: 401,
		},
		{
			Token:      "error",
			StatusCode: 500,
		},
		{
			Token:      "empty",
			StatusCode: 401,
		},
	}

	t.Run("test model user authentication", func(t *testing.T) {
		for _, test := range testCases {
			req, err := http.NewRequest("GET", "/model", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("token", test.Token)

			rr := httptest.NewRecorder()

			model.ServeHTTP(rr, req)

			if status := rr.Code; status != test.StatusCode {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, 405)
			}
		}
	})
}

func TestModelDefaultMethodModel(t *testing.T) {
	//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/4
	methods := []string{
		"POST",
		"OPTIONS",
	}

	t.Run("test model unsupported http method", func(t *testing.T) {
		for _, method := range methods {
			req, err := http.NewRequest(method, "/model", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("token", "abc")

			rr := httptest.NewRecorder()

			model := Model{Auth: AuthTest{}}
			model.ServeHTTP(rr, req)

			if status := rr.Code; status != 405 {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, 405)
			}
		}
	})
}
