package goconf

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"reflect"
)

func decodeJSONRawValue(raw jsontext.Value) reflect.Value {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return reflect.Value{}
	}

	switch trimmed[0] {
	case '"':
		var s metaSeg
		if err := json.Unmarshal(raw, &s); err != nil {
			return reflect.ValueOf(string(raw))
		}

		return reflect.ValueOf(string(s))
	case 't':
		return reflect.ValueOf(true)
	case 'f':
		return reflect.ValueOf(false)
	case 'n':
		return reflect.Value{}
	default:
		return decodeJSONNumber(raw)
	}
}

func decodeJSONNumber(raw jsontext.Value) reflect.Value {
	var i int64
	if err := json.Unmarshal(raw, &i); err == nil {
		return reflect.ValueOf(i)
	}

	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return reflect.ValueOf(f)
	}

	return reflect.ValueOf(string(raw))
}
