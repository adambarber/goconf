package goconf

import (
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

// elemType unwraps pointers, slices, and arrays to the element/container
// type that key segments match against (array-of-tables elements, etc.).
func elemType(t reflect.Type) reflect.Type {
	for {
		k := t.Kind()
		if k == reflect.Pointer || k == reflect.Slice || k == reflect.Array {
			t = t.Elem()
			continue
		}

		return t
	}
}

// isDynamicMapField reports whether f is a struct-tagged dynamic map field
// (arbitrary child keys accepted).
func isDynamicMapField(f *reflect.StructField) bool {
	return isDynamicField(f) && elemType(f.Type).Kind() == reflect.Map
}

// isDynamicValueMap reports whether the current dynamic value context is
// itself a map requiring further dynamic key absorption.
func isDynamicValueMap(t reflect.Type) bool {
	t = elemType(t)
	return t.Kind() == reflect.Map
}

// mapValueType returns the element type for a map-like field, or the
// dereferenced type when not a map.
func mapValueType(t reflect.Type) reflect.Type {
	t = elemType(t)
	if t.Kind() == reflect.Map {
		return t.Elem()
	}

	return t
}

// dynamicValueType returns the value type of a dynamic map for schema-guided
// descent into structured dynamic values.
func dynamicValueType(t reflect.Type) reflect.Type { return mapValueType(t) }

// keySegment is one segment of a document key being mapped to a field.
type keySegment string

// fieldMatching matches one segment against a field's canonical config name,
// case-insensitively. Presence canonicalization uses the format-scoped matcher
// instead; every other caller matches the shared first-tag vocabulary.
func fieldMatching(t reflect.Type, seg keySegment) (*reflect.StructField, bool) {
	t = elemType(t)
	if t.Kind() != reflect.Struct {
		return nil, false
	}

	for i := range t.NumField() {
		f := fieldAt(t, i)
		if !f.IsExported() || string(tagPathOf(f)) == "-" {
			continue
		}

		if strings.EqualFold(string(tagPathOf(f)), string(seg)) {
			return f, true
		}
	}

	return nil, false
}

// tomlKeyMap is the decode target of the key walk, keyed by table name.
type tomlKeyMap map[tomlKey]toml.Primitive

// tomlKey is one top-level TOML table name.
type tomlKey string
