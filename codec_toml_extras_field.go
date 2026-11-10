package goconf

import (
	"reflect"
)

func findStructFieldByKey(t reflect.Type, key keySegment) (*reflect.StructField, bool) {
	f, ok := fieldMatching(t, key)
	if !ok || isExtrasField(f) {
		return nil, false
	}

	return f, true
}

func derefValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return v
		}

		v = v.Elem()
	}

	return v
}

func rawChildRV(raw reflect.Value, key keySegment) reflect.Value {
	if raw.Kind() == reflect.Interface && !raw.IsNil() {
		raw = raw.Elem()
	}

	if raw.Kind() != reflect.Map {
		return reflect.Value{}
	}

	v := raw.MapIndex(reflect.ValueOf(string(key)))
	if v.IsValid() && v.Kind() == reflect.Interface && !v.IsNil() {
		return v.Elem()
	}

	return v
}
