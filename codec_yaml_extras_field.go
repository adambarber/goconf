package goconf

import (
	"reflect"
	"strings"
)

func yamlFieldMatches(f *reflect.StructField, key keySegment) bool {
	if !f.IsExported() || isExtrasField(f) {
		return false
	}

	name := string(yamlFieldName(f))
	if name == "-" {
		return false
	}

	return strings.EqualFold(name, string(key)) || strings.EqualFold(f.Name, string(key))
}

func yamlFieldName(f *reflect.StructField) keySegment {
	tag := f.Tag.Get("yaml")
	if tag == "" {
		tag = f.Tag.Get("json")
	}

	if tag == "" {
		return keySegment(f.Name)
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return keySegment(f.Name)
	}

	return keySegment(name)
}

func derefAlloc(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() && v.CanSet() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
	}

	return v
}
