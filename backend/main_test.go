package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testOptions struct {
	method string
	params interface{}
}

func TestAPI(t *testing.T) {
	type to = testOptions
	type m = map[string]interface{}
	user := newUser("name", "name@example.com", "11111111H", "12345678")
	login := m{"unique_id": "11111111H", "password": "12345678"}
	t.Run("Empty site should not be initialized", testEndpoint("/uninitialized", 200))
	t.Run("Uninitialized site should reject registers", testEndpoint("/auth/register", 401, to{method: "POST", params: user}))
	t.Run("Uninitialized site should reject logins", testEndpoint("/auth/login", 401, to{method: "POST", params: login}))

	admin := newUser("admin", "admin@example.com", "22222222J", "12345678")
	election := newElection("election", "borda", time.Now().Add(1*time.Hour), time.Now().Add(2*time.Hour), 2, 5)
	t.Run("Empty site can be initialized", testEndpoint("/initialize", 200, to{method: "POST", params: m{"admin": admin, "election": election}}))
	t.Run("Initialized site should accept registers", testEndpoint("/auth/register", 200, to{method: "POST", params: user}))
	t.Run("Initialized site should accept logins from registered users", testEndpoint("/auth/login", 200, to{method: "POST", params: login}))
}

func newUser(name, email, uniqueID, password string) map[string]interface{} {
	return map[string]interface{}{
		"name":      name,
		"email":     email,
		"unique_id": uniqueID,
		"password":  password,
	}
}

func newElection(name, countType string, start, end time.Time, minCandidates, maxCandidates int) map[string]interface{} {
	return map[string]interface{}{
		"name":           name,
		"start":          start.Format(time.RFC3339Nano),
		"end":            end.Format(time.RFC3339Nano),
		"count_type":     countType,
		"min_candidates": minCandidates,
		"max_candidates": maxCandidates,
	}
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
