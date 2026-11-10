package goconf

import (
	"reflect"
	"strings"
)

func matchJSONStructField(t reflect.Type, key keySegment) (*reflect.StructField, bool) {
	t = derefTargetType(reflect.New(t).Elem())
	if t.Kind() != reflect.Struct {
		return nil, false
	}

	return jsonStructFieldByKey(t, key)
}

func jsonStructFieldByKey(t reflect.Type, key keySegment) (*reflect.StructField, bool) {
	var foldMatch *reflect.StructField

	n := t.NumField()
	for i := range n {
		f := fieldAt(t, i)
		if !jsonFieldEligible(f) {
			continue
		}

		name := string(jsonFieldName(f))
		if name == string(key) {
			return f, true
		}

		if foldMatch == nil && strings.EqualFold(name, string(key)) {
			foldMatch = f
		}
	}

	return foldMatch, foldMatch != nil
}

func jsonFieldEligible(f *reflect.StructField) bool {
	if !f.IsExported() || isExtrasField(f) {
		return false
	}

	return jsonFieldName(f) != "-"
}

func jsonFieldName(f *reflect.StructField) keySegment {
	tag, ok := f.Tag.Lookup("json")
	if !ok {
		return keySegment(f.Name)
	}

	name, _, _ := strings.Cut(tag, ",")
	if name != "" {
		return keySegment(name)
	}

	return keySegment(f.Name)
}
