// TODO validators should clean up values in addition to validate them
// TODO strip blank space, clean email, etc.
package par

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/mail"
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

func None() func(*http.Request) (Values, error) {
	return func(*http.Request) (Values, error) { return make(Values), nil }
}

func Custom(validator func(*http.Request) (interface{}, error)) params {
	return params{kind: "custom", customValidator: validator}
}

func (p params) Int(name string, validators ...func(interface{}) error) params {
	return p.newParam("int", name, validators...)
}

func (p params) String(name string, validators ...func(interface{}) error) params {
	return p.newParam("string", name, validators...)
}

func (p params) File(name string) params {
	return p.newParam("file", name)
}

func (p params) Email(name string) params {
	return p.newParam("string", name, NonEmpty, Email)
}

func (p params) newParam(kind, name string, validators ...func(interface{}) error) params {
	p.valueKinds[name] = kind
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
		case "json":
			return endJsonParams(r, p)
		case "form":
			return endFormParams(r, p)
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
			panic(fmt.Sprintf("unknown value kind %q", kind))
		}
	}
	return vals, nil
}

func endJsonParams(r *http.Request, p params) (Values, error) {
	m, err := getJsonFromBody(r)
	if err != nil {
		return nil, fmt.Errorf("could not decode body: %w", err)
	}
	vals := make(Values)
	for name, kind := range p.valueKinds {
		switch kind {
		case "string":
			v, ok := m[name]
			if !ok {
				return nil, errMissingParameter
			}
			vv, ok := v.(string)
			if !ok {
				return nil, errWrongType
			}
			if err := checkValidators(vv, name, p.validators); err != nil {
				return nil, err
			}
			vals[name] = vv
		case "int":
			v, ok := m[name]
			if !ok {
				return nil, errMissingParameter
			}
			f, ok := v.(float64)
			if !ok {
				return nil, errWrongType
			}
			vv := int(f)
			if err := checkValidators(vv, name, p.validators); err != nil {
				return nil, err
			}
			vals[name] = vv
		default:
			panic(fmt.Sprintf("unknown value kind %q", kind))
		}
	}
	return vals, nil
}

func endFormParams(r *http.Request, p params) (Values, error) {
	vals := make(Values)
	for name, kind := range p.valueKinds {
		switch kind {
		case "string":
			v := r.FormValue(name)
			if err := checkValidators(v, name, p.validators); err != nil {
				return nil, err
			}
			vals[name] = v
		case "file":
			file, handler, err := r.FormFile(name)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			b, err := ioutil.ReadAll(file)
			if err != nil {
				return nil, err
			}

			if handler.Filename == "" || len(b) == 0 {
				return nil, errMissingParameter
			}

			vals[name] = b
			vals[fileNameField(name)] = handler.Filename
		default:
			panic(fmt.Sprintf("unknown value kind %q", kind))
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
	if v == "" {
		return nil, errMissingParameter
	}

	vv, err := strconv.Atoi(v)
	if err != nil {
		return nil, errWrongType
	}
	if err := checkValidators(vv, name, p.validators); err != nil {
		return nil, err
	}

	return vv, nil
}

func checkValidators(v interface{}, name string, validators map[string][]func(interface{}) error) error {
	for i, validator := range validators[name] {
		if err := validator(v); err != nil {
			return fmt.Errorf("validator %d failed on parameter %s: %w", i, name, err)
		}
	}
	return nil
}

func (v Values) Int(name string) int {
	x, ok := v[name]
	if !ok {
		panic(fmt.Sprintf("asked for unknown name %q", name))
	}

	i, ok := x.(int)
	if !ok {
		panic(fmt.Sprintf("asked for wrong type, expected int, got %T", x))
	}

	return i
}

func (v Values) String(name string) string {
	x, ok := v[name]
	if !ok {
		panic(fmt.Sprintf("asked for unknown name %q", name))
	}

	s, ok := x.(string)
	if !ok {
		panic(fmt.Sprintf("asked for wrong type, expected string, got %T", x))
	}

	return s
}

func (v Values) File(name string) ([]byte, string) {
	x, ok := v[name]
	if !ok {
		panic(fmt.Sprintf("asked for unknown name %q", name))
	}

	b, ok := x.([]byte)
	if !ok {
		panic(fmt.Sprintf("asked for wrong type, expected []byte, got %T", b))
	}

	y := v[fileNameField(name)]
	filename, ok := y.(string)
	if !ok {
		panic(fmt.Sprintf("missing filename field for %q", name))
	}

	return b, filename
}

func fileNameField(name string) string {
	return name + ";_;fileNameField"
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

var NonEmpty = MinLength(1)

func MinLength(length int) func(interface{}) error {
	return func(i interface{}) error {
		v, ok := i.(string)
		if !ok {
			return errWrongType
		}
		if len(v) < length {
			return fmt.Errorf("string length is less than %d", length)
		}
		return nil
	}
}

func Email(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return errWrongType
	}

	_, err := mail.ParseAddress(v)
	if err != nil {
		return err
	}

	return nil
}

func getJsonFromBody(r *http.Request) (m map[string]interface{}, err error) {
	if r.Body == nil {
		return nil, errors.New("empty body")
	}

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		return nil, err
	}

	return m, nil
}
