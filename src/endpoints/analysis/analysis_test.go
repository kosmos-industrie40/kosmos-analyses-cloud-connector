// build +unit
package analysis

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/analysis/models"
	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/src/endpoints/auth"
)

// using this varibale to control the behavior of the GetAllAnalysess function

type testAnalysisHandler struct{}

func (testAnalysisHandler) Insert(contract, machine, sensor string, ana models.Analysis) error {
	if contract == "error" {
		return fmt.Errorf("error")
	}
	return nil
}

func (testAnalysisHandler) Query(contract string, resultID int64) (models.Analysis, error) {
	if contract == "error" {
		return models.Analysis{}, fmt.Errorf("error")
	}
	return models.Analysis{}, nil
}

type testResultHandler struct{}

func (t testResultHandler) Get(contract string, query map[string][]string) ([]byte, error) {
	switch contract {
	case "error":
		return nil, fmt.Errorf("error")
	case "one":
		return []byte("[{\"resultId\":1,\"machine\":\"\",\"date\":0}]"), nil
	case "two":
		return []byte("[{\"resultId\":1,\"machine\":\"\",\"date\":0},{\"resultId\":2,\"machine\":\"\",\"date\":0}]"), nil
	default:
		return nil, nil
	}
}

type testAuthHelper struct{}

func (h testAuthHelper) TokenValid(r *http.Request) (bool, error) {
	panic("implement me")
}

func (testAuthHelper) CreateSession(s string, i []string, x []string, t time.Time) error {
	panic("implement me")
}

func (testAuthHelper) DeleteSession(s string) error {
	panic("implement me")
}

func (testAuthHelper) CleanUp() {
	panic("implement me")
}

func (testAuthHelper) IsAuthenticated(r *http.Request, contract string, write bool) (bool, int, error) {
	return true, 0, nil
}

func (testAuthHelper) ContractWriteAccess(r *http.Request) (bool, int, error) {
	return true, 0, nil
}

var (
	aHandler models.AnalysisHandler   = testAnalysisHandler{}
	tHandler models.ResultListHandler = testResultHandler{}
	aHelper  auth.Helper              = testAuthHelper{}
)

var analyses Analysis = analysis{
	analysis: analyseLogic{
		analysisHandler: aHandler,
		resultHandler:   tHandler,
	},
	authHelper: aHelper,
}

var validModel = `
[{
  "$schema": "analysis-formal.json",
  "from": "creator of this message",
  "timestamp": "2020-08-12T15:46:10.821Z",
  "model": {
    "url": "abc",
    "tag": "ab"
  },
  "type": "text",
  "calculated": {
    "message": {
      "machine": "abc",
      "sensor": "134wdsf"
    },
    "received": "2020-08-12T15:47:10.821Z"
  },
  "results": {
    "total": "stop",
    "predict": 80,
    "parts": [
        {
          "machine": "machine1",
          "result": "stop",
          "predict": 90,
          "sensors": [
            {
              "sensor": "sensor1",
              "result": "stop",
              "predict": 100
            }
          ]
        }
      ]
  },
  "signature": "iQJNBAABCAA3FiEEs9+aTrHmy7aElz5+uqd9AltsqzQFAl6VcaIZHGphbi51bnRlcmJyaW5rQGlub3ZleC5kZQAKCRC6p30CW2yrNATdD/9K0PhIDEHu8lJes+JE8R6t0dIaIauQeRHxO4etowXMiOTaLyGJAiAwM0J5bA3ZfmRzD9eEBpsA74/zXNIFX6FZ8W30Eo+M4Iz5zZ6izi9XcUgKYeyVCEz+vg3Y+JKdrwDQMTshnsJhvK01jkp+VS5deWsojE9T6LQjLEEplcCvqdyHxEAZRyGePsgnJE5eJ9Iw96+nTzuL+OZBhQZDzki9W6utIbVHO6lwM3rYsj9xFQCRzimoxj3WV5HIRFuxFEuzw4GpucaAOfQFd5qSF3qIETQz7wUGJoci0dQcXmarKR4Q+zmUZCh+m5Lps8kh3VUg45bTraVZyzXyF+33p5rUEFJRoKR9VUy7+1YafxoXxXyy9Qa8SPUw1klKMEKem9jIl2iUCZh5ws689HxFox1/sUxfN+Zbli2FhXHg+g1XiqlpcEslAVTr24ixcL6d/R0pERytuspVX0xiX2RU9nNl8RDHGG+XWGppnkA4UQPYEkMSMYt9wT+xonWBHL2vGRF1/s6rVrtbnXWOmKq0kXyTflEQDATUn241nDNXMKu2+IZeXpQlMY1bt8mkxn7xAIbKigNa53geYy9IgmJp83kJmBg2Rs4pqie39QSIVmoXqjVrbPWJjzcWQyl03ShUYV8lXJTrYDxn5VhqFz6ifQima1NEiGOVtai4LrRZcJIljw=="
}]
`

func TestAnalysesPost(t *testing.T) {
	testTable := []struct {
		description string
		statusCode  int
		path        string
		data        string
	}{
		{
			"request data not valid",
			400,
			"/analyses",
			"",
		},
		{
			"request data not valid with contract, machine and sensor",
			400,
			"a/analyses/error/c/v",
			"",
		},
		{
			"internal error",
			500,
			"a/analyses/error/c/v",
			validModel,
		},
		{
			"success",
			201,
			"a/analyses/t/c/v",
			validModel,
		},
	}

	for _, test := range testTable {
		t.Run(test.description, func(t *testing.T) {
			req, err := http.NewRequest("POST", test.path, strings.NewReader(test.data))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("token", "abc")

			rr := httptest.NewRecorder()

			analyses.ServeHTTP(rr, req)

			if status := rr.Code; status != test.statusCode {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, test.statusCode)
			}
		})
	}
}

func TestAnalysesGet(t *testing.T) {
	testTable := []struct {
		description string
		statusCode  int
		path        string
		data        string
	}{
		{
			"internal error",
			500,
			"/analysis/error",
			"",
		},
		{
			"bad request",
			400,
			"/analysis",
			"",
		},
		{
			"success but not found",
			200,
			"/analysis/abc",
			"",
		},
		{
			"analysis one result",
			200,
			"/analysis/one",
			"[{\"resultId\":1,\"machine\":\"\",\"date\":0}]",
		},
		{
			"analysis two results",
			200,
			"/analysis/two",
			"[{\"resultId\":1,\"machine\":\"\",\"date\":0},{\"resultId\":2,\"machine\":\"\",\"date\":0}]",
		},
		{
			"internal error",
			500,
			"/analysis/error/",
			"",
		},
		{
			"sucess empty response",
			200,
			"/analyses/abc/",
			"",
		},
		{
			"success one result without machine",
			200,
			"/analysis/one/",
			"[{\"resultId\":1,\"machine\":\"\",\"date\":0}]",
		},
		{
			"success two result without machine",
			200,
			"/analysis/two/",
			"[{\"resultId\":1,\"machine\":\"\",\"date\":0},{\"resultId\":2,\"machine\":\"\",\"date\":0}]",
		},
		{
			"parse result id",
			400,
			"/analysis/error/ab",
			"",
		},
		{
			"internal error",
			500,
			"/analysis/error/432",
			"",
		},
		{
			"success analysis with empty response",
			200,
			"/analysis/abc/430",
			"{\"from\":\"\",\"timestamp\":\"\",\"model\":{\"url\":\"\",\"tag\":\"\"},\"type\":\"\",\"calculated\":{\"message\":{\"machine\":\"\",\"sensor\":\"\"},\"received\":\"\"},\"results\":null,\"signature\":\"\"}",
		},
		{
			"",
			200,
			"/analysis/one/432",
			"{\"from\":\"\",\"timestamp\":\"\",\"model\":{\"url\":\"\",\"tag\":\"\"},\"type\":\"\",\"calculated\":{\"message\":{\"machine\":\"\",\"sensor\":\"\"},\"received\":\"\"},\"results\":null,\"signature\":\"\"}",
		},
	}

	for _, test := range testTable {
		t.Run(test.description, func(t *testing.T) {
			req, err := http.NewRequest("GET", test.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("token", "abc")

			rr := httptest.NewRecorder()

			analyses.ServeHTTP(rr, req)

			if status := rr.Code; status != test.statusCode {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, test.statusCode)
			}

			if rr.Body.String() != test.data {
				t.Errorf("%v\thandler returnes wrong data in body: got %s want %s", test, rr.Body.String(), test.data)
			}

		})
	}

}

func TestAnalysesDefault(t *testing.T) {
	//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/4
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
