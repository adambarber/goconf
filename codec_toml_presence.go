package goconf

import (
	"bytes"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adambarber/gorefuse"
)

// presentTOMLKeys records the dotted field paths of every key present in
// the TOML document: the decoder's key report covers table headers and
// leaf keys, dotted in document order; values are never inspected.
func presentTOMLKeys(data SourceBytes, file FilePath, t reflect.Type) (KeySet, error) {
	keys := KeySet{paths: nil}

	var raw tomlKeyMap

	md, err := toml.NewDecoder(bytes.NewReader(data)).Decode(&raw)
	if err != nil {
		return keys, tomlRefusal(file, errTOMLValueFree(err))
	}

	for _, k := range md.Keys() {
		p, err := canonicalTOMLPath(k, file, t)
		if err != nil {
			return keys, err
		}

		keys = keys.Add(p)
	}

	return keys, nil
}

// canonicalTOMLPath maps one document key through the decode target's
// field vocabulary, absorbing dynamic/extras keys and refusing unknowns.
func canonicalTOMLPath(k toml.Key, file FilePath, t reflect.Type) (FieldPath, error) {
	path := FieldPath("")
	// inDynamic tracks whether we are descending inside a dynamic map's
	// key space: when true, the current segment is an arbitrary map key.
	inDynamic := false

	for _, seg := range k {
		t = elemType(t)

		switch {
		case inDynamic:
			// Arbitrary key inside a dynamic map: accept the segment as
			// the map entry key and descend into the map's value type
			// for schema-guided validation of structured values.
			path = childPath(path, FieldPath(seg))
			t = dynamicValueType(t)
			inDynamic = isDynamicValueMap(t)

		case t.Kind() == reflect.Struct:
			f, ok := fieldMatching(t, keySegment(seg))
			if ok {
				path = childPath(path, tagPathOf(f))
				t = f.Type
				inDynamic = isDynamicMapField(f)

				continue
			}

			if ef, ok := extrasField(t); ok {
				// Fixed+extras: unknown-to-fixed key is captured by extras.
				path = childPath(path, FieldPath(seg))
				t = extrasValueType(ef.Type)
				inDynamic = false

				continue
			}

			return "", refuseAt(ErrUnknownField, file, 0,
				gorefuse.Subject(`"`+k.String()+`"`),
				gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))

		case t.Kind() == reflect.Map:
			// Descended into a map type (e.g. nested dynamic). Any key is valid.
			path = childPath(path, FieldPath(seg))
			t = t.Elem()
			inDynamic = false

		default:
			return "", refuseAt(ErrUnknownField, file, 0,
				gorefuse.Subject(`"`+k.String()+`"`),
				gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
		}
	}

	return path, nil
}

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

// extrasValueType returns the value type captured by an extras map field.
func extrasValueType(t reflect.Type) reflect.Type { return mapValueType(t) }

// keySegment is one segment of a document key being mapped to a field.
type keySegment string

// fieldMatching returns the struct field whose tag or Go name matches
// seg case-insensitively, unexported fields skipped.
func fieldMatching(t reflect.Type, seg keySegment) (*reflect.StructField, bool) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, false
	}

	for i := range t.NumField() {
		f := fieldAt(t, i)
		if !f.IsExported() {
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
