package par

import (
	"net/http"
	"testing"
)

func TestParams(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost?id=123", nil)
	if err != nil {
		t.Errorf("Could not define request: %s", err)
	}

	pf := P("query").Int("id", PositiveInt).End()
	values, err := pf(req)
	if err != nil {
		t.Errorf("Error parsing params: %s.", err)
	}

	id := values.Int("id")
	if id != 123 {
		t.Errorf("Expected 123, but got %d.", id)
	}

	assertPanic(t, func() { values.Int("invalid-name") })
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but got none")
		}
	}()
	f()
}
