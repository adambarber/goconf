package goconf

import (
	"bytes"
	"reflect"
)

func jsonFieldMatches(f *reflect.StructField, key keySegment) bool {
	if !f.IsExported() || isExtrasField(f) {
		return false
	}

	name := jsonFieldName(f)
	if name == "-" {
		return false
	}

	return string(name) == string(key)
}

func jsonFieldName(f *reflect.StructField) keySegment {
	tag := f.Tag.Get("json")
	if tag == "" {
		return keySegment(f.Name)
	}

	name, _, _ := bytes.Cut([]byte(tag), []byte(","))
	if len(name) == 0 {
		return keySegment(f.Name)
	}

	return keySegment(string(name))
}
