package par

import (
	"net/http"
	"strconv"
	"testing"
)

func TestInt(t *testing.T) {
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

func TestCustom(t *testing.T) {
	type p struct {
		a int
		b string
	}

	validator := func(r *http.Request) (interface{}, error) {
		aq := r.URL.Query().Get("a")
		a, _ := strconv.Atoi(aq)
		b := r.URL.Query().Get("b")
		return p{a: a, b: b}, nil
	}

	req, err := http.NewRequest("GET", "http://localhost?a=123&b=asd", nil)
	if err != nil {
		t.Errorf("Could not define request: %s", err)
	}

	pf := Custom(validator).End()
	values, err := pf(req)
	if err != nil {
		t.Errorf("Error parsing params: %s.", err)
	}

	x := values.Custom()
	y, ok := x.(p)
	if !ok {
		t.Errorf("Wrong type returned by custom: %T", x)
	}

	if y.a != 123 || y.b != "asd" {
		t.Errorf("Expected %v, but got %v.", p{a: 123, b: "asd"}, y)
	}
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but got none")
		}
	}()
	f()
}
