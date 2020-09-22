// build +unit
package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/database"
)

type AuthTest struct{}

func (a AuthTest) User(token string) (string, error) {
	switch token {
	default:
		return token, nil
	case "error":
		return "", fmt.Errorf("error")
	case "empty":
		return "", nil

	}
}

func (a AuthTest) Authentication(_ database.Postgres) {}

func (a AuthTest) Login(user, password string) (string, error) {
	if user == "error" {
		return "", fmt.Errorf("error")
	}
	return password, nil
}

func (a AuthTest) Logout(token string) error {
	switch token {
	case "error":
		return fmt.Errorf("%s", token)
	}
	return nil
}

type testCases struct {
	Token      string
	StatusCode int
	Expected   string
	Method     string
}

func TestUsedMethodWithToken(t *testing.T) {
	//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/4
	testsGetUser := []testCases{
		{Method: "GET", Token: "", StatusCode: 204, Expected: ""},
		{Method: "GET", Token: "error", StatusCode: 500, Expected: ""},
		{Method: "GET", Token: "empty", StatusCode: 204, Expected: ""},
		{Method: "GET", Token: "user", StatusCode: 200, Expected: "{\"user\":\"user\"}"},
		{Method: "DELETE", Token: "", StatusCode: 201, Expected: ""},
		{Method: "DELETE", Token: "error", StatusCode: 500, Expected: ""},
		{Method: "DELETE", Token: "empty", StatusCode: 201, Expected: ""},
		{Method: "DELETE", Token: "user", StatusCode: 201, Expected: ""},
	}

	auth := Auth{Auth: AuthTest{}}

	t.Run("user test actions on authenticated user", func(t *testing.T) {
		for _, tests := range testsGetUser {
			req, err := http.NewRequest(tests.Method, "/auth", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("token", tests.Token)

			rr := httptest.NewRecorder()

			auth.ServeHTTP(rr, req)

			if status := rr.Code; status != tests.StatusCode {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, tests.StatusCode)
			}

			if rr.Body.String() != tests.Expected {
				t.Errorf("handler return wrong return string: get %s want %s", rr.Body.String(), tests.Expected)
			}

		}
	})
}

func TestDefaultUnexpectedHttMethod(t *testing.T) {
	//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/4
	methods := []string{
		"PUT",
		"OPTIONS",
	}

	for _, method := range methods {
		req, err := http.NewRequest(method, "/auth", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()

		auth := Auth{Auth: AuthTest{}}
		auth.ServeHTTP(rr, req)

		t.Run("test user unsupported http methods", func(t *testing.T) {
			if status := rr.Code; status != 405 {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, 405)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	//TODO https://gitlab.inovex.de/proj-kosmos/kosmos-analyses-cloud-connector/-/issues/4
	loginUser := []struct {
		StatusCode int
		Data       string
		Expected   string
	}{
		{
			StatusCode: 400,
			Data:       "user",
			Expected:   "",
		},
		{
			StatusCode: 200,
			Data:       "{\"user\":\"user\", \"password\":\"password\"}",
			Expected:   "{\"token\":\"password\"}",
		},
		{
			StatusCode: 500,
			Data:       "{\"user\":\"error\", \"password\":\"password\"}",
			Expected:   "",
		},
	}
	auth := Auth{Auth: AuthTest{}}
	for _, tests := range loginUser {
		req, err := http.NewRequest("POST", "/auth", strings.NewReader(tests.Data))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		auth.ServeHTTP(rr, req)

		t.Run("test user login", func(t *testing.T) {
			if status := rr.Code; status != tests.StatusCode {
				t.Errorf("handler returnes wrong status code: got %d want %d", status, tests.StatusCode)
			}

			if rr.Body.String() != tests.Expected {
				t.Errorf("handler return wrong return string: get %s want %s", rr.Body.String(), tests.Expected)
			}
		})

	}
}
