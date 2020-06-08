// build +unit
package endpoints

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/models"
)

// using this varibale to control the behavior of the GetAllAnalysess function

type testAnalysis struct{}

func (a testAnalysis) Analyses(d database.Postgres) {}

func (a testAnalysis) InsertResult(id, contract, sensor string, data []models.UploadResult) error {
	switch id {
	case "error":
		return fmt.Errorf("error")
	default:
		return nil
	}
}

func (a testAnalysis) GetSpecificResult(id, contract string) ([]byte, error) {
	switch id {
	case "error":
		return nil, fmt.Errorf("error")
	case "one":
		return []byte("one"), nil
	default:
		return nil, nil
	}
}

func (a testAnalysis) GetResultSet(id string, query map[string][]string) ([]models.ResultList, error) {
	switch id {
	case "error":
		return nil, fmt.Errorf("error")
	case "one":
		return []models.ResultList{{Id: 1}}, nil
	case "two":
		return []models.ResultList{{Id: 1}, {Id: 2}}, nil
	default:
		return nil, nil
	}
}

var analyses Analyses = Analyses{Auth: AuthTest{}, Analyses: testAnalysis{}}

func TestAnalysesPost(t *testing.T) {

	testCases := []struct {
		StatusCode int
		Path       string
		Data       string
	}{
		{
			StatusCode: 400,
			Path:       "/analyses",
			Data:       "",
		},
		{
			StatusCode: 400,
			Path:       "a/analyses/error/c/v",
			Data:       "",
		},
		{
			StatusCode: 500,
			Path:       "a/analyses/error/c/v",
			Data:       "[{\"Date\": 123, \"From\": \"ich\", \"Type\": \"du\"}]",
		},
		{
			StatusCode: 201,
			Path:       "a/analyses/t/c/v",
			Data:       "[{\"Date\": 123, \"From\": \"ich\", \"Type\": \"du\"}]",
		},
	}

	t.Run("test analyses post cases", func(t *testing.T) {
		for _, test := range testCases {
			req, err := http.NewRequest("POST", test.Path, strings.NewReader(test.Data))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("token", "abc")

			rr := httptest.NewRecorder()

			analyses.ServeHTTP(rr, req)

			if status := rr.Code; status != test.StatusCode {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, test.StatusCode)
			}

		}
	})
}

func TestAnalysesGet(t *testing.T) {
	testCases := []struct {
		StatusCode int
		Path       string
		Data       string
	}{
		{
			StatusCode: 500,
			Path:       "/analyses/error",
			Data:       "",
		},
		{
			StatusCode: 400,
			Path:       "/analyses",
			Data:       "",
		},
		{
			StatusCode: 200,
			Path:       "/analyses/abc",
			Data:       "",
		},
		{
			StatusCode: 200,
			Path:       "/analyses/one",
			Data:       "[{\"resultId\":1,\"machine\":\"\",\"date\":0}]",
		},
		{
			StatusCode: 200,
			Path:       "/analyses/two",
			Data:       "[{\"resultId\":1,\"machine\":\"\",\"date\":0},{\"resultId\":2,\"machine\":\"\",\"date\":0}]",
		},
		{
			StatusCode: 500,
			Path:       "/analyses/error/",
			Data:       "",
		},
		{
			StatusCode: 200,
			Path:       "/analyses/abc/",
			Data:       "",
		},
		{
			StatusCode: 200,
			Path:       "/analyses/one/",
			Data:       "[{\"resultId\":1,\"machine\":\"\",\"date\":0}]",
		},
		{
			StatusCode: 200,
			Path:       "/analyses/two/",
			Data:       "[{\"resultId\":1,\"machine\":\"\",\"date\":0},{\"resultId\":2,\"machine\":\"\",\"date\":0}]",
		},
		{
			StatusCode: 500,
			Path:       "/analyses/error/ab",
			Data:       "",
		},
		{
			StatusCode: 200,
			Path:       "/analyses/abc/ab",
			Data:       "",
		},
		{
			StatusCode: 200,
			Path:       "/analyses/one/ab",
			Data:       "one",
		},
	}

	t.Run("test analyses get cases", func(t *testing.T) {
		for _, test := range testCases {
			req, err := http.NewRequest("GET", test.Path, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("token", "abc")

			rr := httptest.NewRecorder()

			analyses.ServeHTTP(rr, req)

			if status := rr.Code; status != test.StatusCode {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, test.StatusCode)
			}

			if rr.Body.String() != test.Data {
				t.Errorf("%v\thandler returnes wrong data in body: got %s want %s", test, rr.Body.String(), test.Data)
			}

		}
	})

}

func TestAnalysesUserAuth(t *testing.T) {
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

	req, err := http.NewRequest("GET", "/analyses", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("test analyses user authentication", func(t *testing.T) {
		for _, test := range testCases {
			req.Header.Set("token", test.Token)

			rr := httptest.NewRecorder()

			analyses.ServeHTTP(rr, req)

			if status := rr.Code; status != test.StatusCode {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, test.StatusCode)
			}

		}
	})
}

func TestAnalysesDefault(t *testing.T) {
	options := []string{
		"OPTIONS",
		"PUT",
		"TRACE",
	}

	t.Run("test default http methods", func(t *testing.T) {
		for _, test := range options {
			req, err := http.NewRequest(test, "/analyses", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("token", "abc")

			rr := httptest.NewRecorder()

			analyses.ServeHTTP(rr, req)

			if status := rr.Code; status != 405 {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, 405)
			}

		}
	})
}
