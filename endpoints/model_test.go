package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
)

type mod struct{}

func (m mod) Model(_ database.Postgres) {}

func (m mod) GetModel(contract string) ([]models.Model, error) {
	switch contract {
	default:
		return nil, nil
	case "error":
		return nil, fmt.Errorf("error")
	case "empty":
		return []models.Model{}, nil
	case "two":
		ret := []models.Model{
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
		ret := []models.Model{
			{
				Tag: "tag",
				Url: "url",
			},
		}
		return ret, nil
	}
}

func (m mod) UpdateModel(contract string, model models.UpdateModelState) error {
	if contract == "error" {
		return fmt.Errorf("error")
	}
	return nil
}

var model Model = Model{Auth: AuthTest{}, Model: mod{}}

func TestModelUpdate(t *testing.T) {
	mod, err := json.Marshal(models.UpdateModelState{})
	if err != nil {
		t.Fatal(t)
	}
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
}

func TestModelGet(t *testing.T) {
	testCases := []struct {
		StatusCode int
		Expected   []models.Model
		Path       string
	}{
		{
			StatusCode: 400,
			Path:       "/",
			Expected:   []models.Model{},
		},
		{
			StatusCode: 500,
			Path:       "/contract/error",
			Expected:   []models.Model{},
		},
		{
			StatusCode: 200,
			Path:       "/contract/empty",
			Expected:   []models.Model{},
		},
		{
			StatusCode: 200,
			Path:       "/contract/two",
			Expected: []models.Model{
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
			Expected: []models.Model{
				{
					Tag: "tag",
					Url: "url",
				},
			},
		},
	}

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
				t.Errorf("handler returnes wrong body response: got %s want %s", string(rr.Body.Bytes()), "")
			}
			continue
		}

		ex, err := json.Marshal(test.Expected)
		if err != nil {
			t.Fatal(err)
		}

		if string(ex) != rr.Body.String() {
			t.Errorf("handler returnes wrong body response: got %s want %s", string(rr.Body.Bytes()), string(ex))
		}
	}
}

func TestModelAuth(t *testing.T) {
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
}

func TestModelDefaultMethodModel(t *testing.T) {
	methods := []string{
		"POST",
		"OPTIONS",
	}
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
}
