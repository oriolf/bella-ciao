package par

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

var (
	errMissingParameter = errors.New("missing parameter")
	errWrongType        = errors.New("wrong type")
)

type params struct {
	kind            string
	valueKinds      map[string]string
	validators      map[string][]func(interface{}) error
	customValue     interface{}
	customValidator func(*http.Request) (interface{}, error)
}

type Values map[string]interface{}

func P(kind string) params {
	return params{
		kind:       kind,
		valueKinds: make(map[string]string),
		validators: make(map[string][]func(interface{}) error),
	}
}

func Custom(validator func(*http.Request) (interface{}, error)) params {
	return params{kind: "custom", customValidator: validator}
}

func (p params) Int(name string, validators ...func(interface{}) error) params {
	p.valueKinds[name] = "int"
	for _, v := range validators {
		p.validators[name] = append(p.validators[name], v)
	}
	return p
}

func (p params) End() func(*http.Request) (Values, error) {
	return func(r *http.Request) (Values, error) {
		switch p.kind {
		case "query":
			return endQueryParams(r, p)
		case "custom":
			return endCustomParams(r, p)
		}
		panic("unknown params kind!")
	}
}

func endQueryParams(r *http.Request, p params) (Values, error) {
	vals := make(Values)
	for name, kind := range p.valueKinds {
		switch kind {
		case "int":
			v, err := getQueryInt(r, p, name)
			if err != nil {
				return nil, err
			}
			vals[name] = v
		default:
			panic("unknown value kind!")
		}
	}
	return vals, nil
}

func endCustomParams(r *http.Request, p params) (Values, error) {
	val, err := p.customValidator(r)
	if err != nil {
		return nil, err
	}
	return Values(map[string]interface{}{"custom": val}), nil
}

func getQueryInt(r *http.Request, p params, name string) (interface{}, error) {
	v := r.URL.Query().Get(name)
	vv, err := strconv.Atoi(v)
	if err != nil {
		return nil, errMissingParameter
	}
	for i, validator := range p.validators[name] {
		if err := validator(vv); err != nil {
			return nil, fmt.Errorf("validator %d failed on parameter %s: %w", i, name, err)
		}
	}

	return vv, nil
}

func (v Values) Int(name string) int {
	x, ok := v[name]
	if !ok {
		panic("asked for unknown name!")
	}

	i, ok := x.(int)
	if !ok {
		panic("asked for wrong type!")
	}

	return i
}

func (v Values) Custom() interface{} {
	return v["custom"]
}

func PositiveInt(i interface{}) error {
	v, ok := i.(int)
	if !ok {
		return errWrongType
	}
	if v <= 0 {
		return errors.New("int is less or equal to 0")
	}

	return nil
}
