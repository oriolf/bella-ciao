package par

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"
)

var (
	errMissingParameter = errors.New("missing parameter")
	errWrongType        = errors.New("wrong type")
)

type params struct {
	kind            string
	valueKinds      map[string]string
	validators      map[string][]func(interface{}) (interface{}, error)
	customValue     interface{}
	customValidator func(*http.Request) (interface{}, error)
	validateFunc    func(Values) error
	subParams       map[string]func(map[string]interface{}) (Values, error)
}

type Values map[string]interface{}

func P(kind string) params {
	return params{
		kind:       kind,
		valueKinds: make(map[string]string),
		validators: make(map[string][]func(interface{}) (interface{}, error)),
		subParams:  make(map[string]func(map[string]interface{}) (Values, error)),
	}
}

func None() func(*http.Request) (Values, error) {
	return func(*http.Request) (Values, error) { return make(Values), nil }
}

func Custom(validator func(*http.Request) (interface{}, error)) params {
	return params{kind: "custom", customValidator: validator}
}

func (p params) Int(name string, validators ...func(interface{}) (interface{}, error)) params {
	return p.newParam("int", name, validators...)
}

func (p params) String(name string, validators ...func(interface{}) (interface{}, error)) params {
	return p.newParam("string", name, validators...)
}

func (p params) File(name string) params {
	return p.newParam("file", name)
}

func (p params) Email(name string) params {
	return p.newParam("string", name, NonEmpty, Email)
}

func (p params) Time(name string, validators ...func(interface{}) (interface{}, error)) params {
	return p.newParam("time", name, validators...)
}

func (p params) JSON(name string, x func(map[string]interface{}) (Values, error)) params {
	p.valueKinds[name] = "json"
	p.subParams[name] = x
	return p
}

func (p params) newParam(kind, name string, validators ...func(interface{}) (interface{}, error)) params {
	p.valueKinds[name] = kind
	for _, v := range validators {
		p.validators[name] = append(p.validators[name], v)
	}
	return p
}

func (p params) ValidateFunc(f func(Values) error) params {
	p.validateFunc = f
	return p
}

func (p params) End() func(*http.Request) (Values, error) {
	return func(r *http.Request) (v Values, err error) {
		switch p.kind {
		case "query":
			v, err = p.endQueryParams(r)
		case "json":
			v, err = p.endJsonParams(r)
		case "form":
			v, err = p.endFormParams(r)
		case "custom":
			v, err = p.endCustomParams(r)
		default:
			panic("unknown params kind!")
		}
		if err != nil {
			return nil, err
		}
		if p.validateFunc != nil {
			err = p.validateFunc(v)
		}
		return v, err
	}
}

func (p params) EndJSON() func(map[string]interface{}) (Values, error) {
	return func(m map[string]interface{}) (v Values, err error) {
		switch p.kind {
		case "json":
			v, err = p.endJsonParamsAux(m)
		default:
			panic("unknown params kind!")
		}
		if err != nil {
			return nil, err
		}
		if p.validateFunc != nil {
			err = p.validateFunc(v)
		}
		return v, err
	}
}

func (p params) endQueryParams(r *http.Request) (Values, error) {
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

func (p params) endJsonParams(r *http.Request) (Values, error) {
	m, err := getJsonFromBody(r)
	if err != nil {
		return nil, fmt.Errorf("could not decode body: %w", err)
	}
	return p.endJsonParamsAux(m)
}

func (p params) endJsonParamsAux(m map[string]interface{}) (Values, error) {
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
			res, err := checkValidators(vv, name, p.validators)
			if err != nil {
				return nil, err
			}
			vals[name] = res
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
			res, err := checkValidators(vv, name, p.validators)
			if err != nil {
				return nil, err
			}
			vals[name] = res
		case "time":
			v, ok := m[name]
			if !ok {
				return nil, errMissingParameter
			}
			t, ok := v.(string)
			if !ok {
				return nil, errWrongType
			}
			tt, err := time.Parse(time.RFC3339Nano, t)
			if err != nil {
				return nil, err
			}
			res, err := checkValidators(tt, name, p.validators)
			if err != nil {
				return nil, err
			}
			vals[name] = res
		case "json":
			x, ok := m[name]
			if !ok {
				return nil, errMissingParameter
			}

			v, ok := x.(map[string]interface{})
			if !ok {
				return nil, errWrongType
			}

			f, ok := p.subParams[name]
			if !ok {
				panic(fmt.Sprintf("unknown subparams %q", name))
			}

			vv, err := f(v)
			if err != nil {
				return nil, err
			}

			vals[name] = vv
		default:
			panic(fmt.Sprintf("unknown value kind %q", kind))
		}
	}
	return vals, nil
}

func (p params) endFormParams(r *http.Request) (Values, error) {
	vals := make(Values)
	for name, kind := range p.valueKinds {
		switch kind {
		case "string":
			v := r.FormValue(name)
			res, err := checkValidators(v, name, p.validators)
			if err != nil {
				return nil, err
			}
			vals[name] = res
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

func (p params) endCustomParams(r *http.Request) (Values, error) {
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
	res, err := checkValidators(vv, name, p.validators)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func checkValidators(v interface{}, name string, validators map[string][]func(interface{}) (interface{}, error)) (x interface{}, err error) {
	for i, validator := range validators[name] {
		v, err = validator(v)
		if err != nil {
			return v, fmt.Errorf("validator %d failed on parameter %s: %w", i, name, err)
		}
	}
	return v, nil
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

func (v Values) Time(name string) time.Time {
	x, ok := v[name]
	if !ok {
		panic(fmt.Sprintf("asked for unknown name %q", name))
	}

	s, ok := x.(time.Time)
	if !ok {
		panic(fmt.Sprintf("asked for wrong type, expected time.Time, got %T", x))
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

func (v Values) Values(name string) Values {
	x, ok := v[name]
	if !ok {
		panic(fmt.Sprintf("asked for unknown name %q", name))
	}

	vv, ok := x.(Values)
	if !ok {
		panic(fmt.Sprintf("asked for wrong type, expected Values, got %T", x))
	}

	return vv
}

func fileNameField(name string) string {
	return name + ";_;fileNameField"
}

func (v Values) Custom() interface{} {
	return v["custom"]
}

func PositiveInt(i interface{}) (interface{}, error) {
	v, ok := i.(int)
	if !ok {
		return i, errWrongType
	}
	if v <= 0 {
		return i, errors.New("int is less or equal to 0")
	}

	return v, nil
}

var NonEmpty = MinLength(1)

func MinLength(length int) func(interface{}) (interface{}, error) {
	return func(i interface{}) (interface{}, error) {
		v, ok := i.(string)
		if !ok {
			return i, errWrongType
		}
		v = strings.TrimSpace(v)
		if len(v) < length {
			return i, fmt.Errorf("string length is less than %d", length)
		}
		return v, nil
	}
}

func Email(i interface{}) (interface{}, error) {
	v, ok := i.(string)
	if !ok {
		return i, errWrongType
	}

	addr, err := mail.ParseAddress(v)
	if err != nil {
		return i, err
	}
	v = addr.Address

	return v, nil
}

func StringIn(l []string) func(interface{}) (interface{}, error) {
	return func(i interface{}) (interface{}, error) {
		v, ok := i.(string)
		if !ok {
			return i, errWrongType
		}

		if !stringInSlice(v, l) {
			return i, fmt.Errorf("value %s not in allowed values list", v)
		}

		return v, nil
	}
}

func StringValidates(m map[string]func(string) error) func(interface{}) (interface{}, error) {
	return func(i interface{}) (interface{}, error) {
		v, ok := i.(string)
		if !ok {
			return i, errWrongType
		}

		var names []string
		for name, f := range m {
			names = append(names, name)
			if err := f(v); err == nil {
				return v, nil
			}
		}

		return v, fmt.Errorf("string %q did not validate any format %v", v, names)
	}
}

func stringInSlice(s string, l []string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

func NonZeroTime(i interface{}) (interface{}, error) {
	v, ok := i.(time.Time)
	if !ok {
		return i, errWrongType
	}

	if v.IsZero() {
		return i, errors.New("time should not be zero")
	}

	return v, nil
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
