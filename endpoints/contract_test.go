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
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
)

// using this varibale to control the behavior of the GetAllContracts function
var getAllContracts string = "not error"

type cont struct{}

func (c cont) GetContract(id string) (models.Contract, error) {
	switch id {
	default:
		return models.Contract{}, nil
	case "error":
		return models.Contract{}, fmt.Errorf("error")
	}
}

func (c cont) GetAllContracts() ([]string, error) {
	switch getAllContracts {
	default:
		return nil, nil
	case "error":
		return nil, fmt.Errorf("")
	case "one":
		return []string{"one"}, nil
	case "two":
		return []string{"one", "two"}, nil
	case "emtpy":
		return []string{""}, nil
	}
}

func (c cont) InsertContract(data models.Contract) error {
	switch data.ContractId {
	case "error":
		return fmt.Errorf("error")
	default:
		return nil
	}
}

func (c cont) Contract(database.Postgres) {}

func (c cont) DeleteContract(id string) error {
	switch id {
	default:
		return nil
	case "error":
		return fmt.Errorf("error")
	}
}

var contract Contract = Contract{Auth: AuthTest{}, Contract: cont{}}

func TestContractPost(t *testing.T) {
	errorCase := models.Contract{ContractId: "error"}
	successCase := models.Contract{ContractId: "success"}

	errorCaseBytes, err := json.Marshal(errorCase)
	if err != nil {
		t.Fatal(err)
	}
	successCaseBytes, err := json.Marshal(successCase)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		StatusCode int
		Path       string
		Data       string
	}{
		{
			StatusCode: 201,
			Path:       "/contract",
			Data:       string(successCaseBytes),
		},
		{
			StatusCode: 500,
			Path:       "/contract",
			Data:       string(errorCaseBytes),
		},
	}

	for _, test := range testCases {
		req, err := http.NewRequest("POST", test.Path, strings.NewReader(test.Data))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("token", "abc")

		rr := httptest.NewRecorder()

		contract.ServeHTTP(rr, req)

		if status := rr.Code; status != test.StatusCode {
			t.Errorf("handler returnes wrong status code: got %d want %d", status, test.StatusCode)
		}

	}
}

func TestContractGetDelete(t *testing.T) {
	testCases := []struct {
		StatusCode int
		Path       string
		Data       string
		Method     string
		variable   string
	}{
		{
			StatusCode: 400,
			Path:       "/contract",
			Data:       "",
			variable:   "",
			Method:     "DELETE",
		},
		{
			StatusCode: 500,
			Path:       "/contract/error",
			Data:       "",
			Method:     "DELETE",
			variable:   "",
		},
		{
			StatusCode: 201,
			Path:       "/contract/abc",
			Data:       "",
			Method:     "DELETE",
			variable:   "",
		},
		{
			StatusCode: 400,
			Path:       "/contract/abc/d/e",
			Data:       "",
			Method:     "GET",
			variable:   "",
		},
		{
			StatusCode: 200,
			Path:       "/contract",
			Data:       "",
			Method:     "GET",
			variable:   "empty",
		},
		{
			StatusCode: 500,
			Path:       "/contract",
			Data:       "",
			Method:     "GET",
			variable:   "error",
		},
		{
			StatusCode: 200,
			Path:       "/contract",
			Data:       "[\"one\"]",
			Method:     "GET",
			variable:   "one",
		},
		{
			StatusCode: 200,
			Path:       "/contract",
			Data:       "[\"one\",\"two\"]",
			Method:     "GET",
			variable:   "two",
		},
		{
			StatusCode: 200,
			Path:       "/contract",
			Data:       "",
			Method:     "GET",
			variable:   "empty",
		},
		{
			StatusCode: 500,
			Path:       "/contract/",
			Data:       "",
			Method:     "GET",
			variable:   "error",
		},
		{
			StatusCode: 200,
			Path:       "/contract/",
			Data:       "[\"one\"]",
			Method:     "GET",
			variable:   "one",
		},
		{
			StatusCode: 200,
			Path:       "/contract/",
			Data:       "[\"one\",\"two\"]",
			Method:     "GET",
			variable:   "two",
		},
		{
			StatusCode: 200,
			Path:       "/contract/",
			Data:       "",
			Method:     "GET",
			variable:   "empty",
		},
		{
			StatusCode: 500,
			Path:       "/contract/error",
			Data:       "",
			Method:     "GET",
			variable:   "",
		},
		{
			StatusCode: 200,
			Path:       "/contract/abc",
			Data:       "{\"modelsCloud\":null,\"modelsEdge\":null,\"contractId\":\"\",\"machines\":null}",
			Method:     "GET",
			variable:   "",
		},
	}

	for _, test := range testCases {
		getAllContracts = test.variable
		req, err := http.NewRequest(test.Method, test.Path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("token", "abc")

		rr := httptest.NewRecorder()

		contract.ServeHTTP(rr, req)

		if status := rr.Code; status != test.StatusCode {
			t.Errorf("handler returnes wrong status code: got %d want %d", status, test.StatusCode)
		}

		if rr.Body.String() != test.Data {
			t.Errorf("%v\thandler returnes wrong data in body: got %s want %s", test, rr.Body.String(), test.Data)
		}

	}

}

func TestContractUserAuth(t *testing.T) {
	testCases := []struct {
		StatusCode int
		Token      string
	}{
		{
			StatusCode: 401,
			Token:      "",
		},
		{
			StatusCode: 500,
			Token:      "error",
		},
		{
			StatusCode: 401,
			Token:      "empty",
		},
	}

	req, err := http.NewRequest("GET", "/contract", nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range testCases {
		req.Header.Set("token", test.Token)

		rr := httptest.NewRecorder()

		contract.ServeHTTP(rr, req)

		if status := rr.Code; status != test.StatusCode {
			t.Errorf("handler returnes wrong status code: got %d want %d", status, test.StatusCode)
		}

	}
}

func TestContractDefault(t *testing.T) {
	options := []string{
		"OPTIONS",
		"PUT",
		"TRACE",
	}

	for _, test := range options {
		req, err := http.NewRequest(test, "/contract", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("token", "abc")

		rr := httptest.NewRecorder()

		contract.ServeHTTP(rr, req)

		if status := rr.Code; status != 405 {
			t.Errorf("handler returnes wrong status code: got %d want %d", status, 405)
		}

	}
}
