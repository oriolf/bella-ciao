package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPI(t *testing.T) {
	t.Run("Empty site should not be initialized", testEndpoint("/uninitialized", 200))
}

func testEndpoint(path string, expectedCode int) func(*testing.T) {
	return func(t *testing.T) {
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatalf("Could not create request for endpoint %q. Error: %s\n", path, err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(appHandlers[path])
		handler.ServeHTTP(rr, req)
		if rr.Code != expectedCode {
			t.Errorf("Expected code %v testing endpoint %q, but got %v.", expectedCode, path, rr.Code)
		}
	}
}
