package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testOptions struct {
	method string
	params interface{}
}

func TestAPI(t *testing.T) {
	type to = testOptions
	type m = map[string]interface{}
	user := m{
		"name":      "name",
		"email":     "example@example.com",
		"unique_id": "11111111H",
		"password":  "12345678",
	}
	login := m{
		"unique_id": "11111111H",
		"password":  "12345678",
	}
	t.Run("Empty site should not be initialized", testEndpoint("/uninitialized", 200))
	t.Run("Uninitialized site should reject registers", testEndpoint("/auth/register", 401, to{method: "POST", params: user}))
	t.Run("Uninitialized site should reject logins", testEndpoint("/auth/login", 401, to{method: "POST", params: login}))
}

func testEndpoint(path string, expectedCode int, options ...testOptions) func(*testing.T) {
	i := 0
	return func(t *testing.T) {
		i++
		method := "GET"
		var params interface{}
		if len(options) > 0 {
			op := options[0]
			if op.method != "" {
				method = op.method
			}

			if op.params != nil {
				params = op.params
			}
		}

		var body io.Reader
		if params != nil {
			b, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("[%d] Could not marshal params for endpoint %q. Error: %s", i, path, err)
			}
			body = bytes.NewReader(b)
		}

		req, err := http.NewRequest(method, path, body)
		if err != nil {
			t.Fatalf("[%d] Could not create request for endpoint %q. Error: %s\n", i, path, err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(appHandlers[path])
		handler.ServeHTTP(rr, req)
		if rr.Code != expectedCode {
			t.Errorf("[%d] Expected code %v testing endpoint %q, but got %v.", i, expectedCode, path, rr.Code)
		}
	}
}
