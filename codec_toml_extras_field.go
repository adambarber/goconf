package goconf

import (
	"reflect"

	"github.com/BurntSushi/toml"
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

func rawLookupRV(raw reflect.Value, path toml.Key) reflect.Value {
	cur := raw
	for _, p := range path {
		cur = rawChildRV(cur, keySegment(p))
	}

	if cur.IsValid() && cur.Kind() == reflect.Interface && !cur.IsNil() {
		return cur.Elem()
	}

	return cur
}

func mapGetStringKey(m reflect.Value, key keySegment) reflect.Value {
	if m.Kind() != reflect.Map {
		return reflect.Value{}
	}

	return m.MapIndex(reflect.ValueOf(string(key)))
}
