package goconf

import (
	"reflect"
	"strings"
)

// goconfTag is the parsed goconf struct tag for one field.
type goconfTag struct {
	dynamic bool
	extras  bool
}

// parseGoconfTag parses the `goconf` struct tag.
func parseGoconfTag(f *reflect.StructField) goconfTag {
	tag := f.Tag.Get("goconf")
	if tag == "" {
		return goconfTag{dynamic: false, extras: false}
	}

	out := goconfTag{dynamic: false, extras: false}

	for part := range strings.SplitSeq(tag, ",") {
		switch strings.TrimSpace(part) {
		case "dynamic":
			out.dynamic = true
		case "extras", "rest":
			out.extras = true
		default:
		}
	}

	return out
}

// isDynamicField reports whether the field declares a dynamic-key absorbing map.
func isDynamicField(f *reflect.StructField) bool {
	return parseGoconfTag(f).dynamic
}

// isExtrasField reports whether the field captures extra/unknown keys for its struct.
func isExtrasField(f *reflect.StructField) bool {
	return parseGoconfTag(f).extras
}

// extrasField returns the extras-capture field of struct type t, if any.
func extrasField(t reflect.Type) (reflect.StructField, bool) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	var zero reflect.StructField

	if t.Kind() != reflect.Struct {
		return zero, false
	}

	for f := range t.Fields() {
		if !f.IsExported() {
			continue
		}

		if isExtrasField(&f) {
			return f, true
		}
	}

	return zero, false
}

// hasExtrasInType reports whether t or any reachable field type contains
// a struct with an extras-capture field.
func hasExtrasInType(t reflect.Type) bool {
	return hasExtrasInTypeSeen(t, map[reflect.Type]bool{})
}

func hasExtrasInTypeSeen(t reflect.Type, seen map[reflect.Type]bool) bool {
	if t == nil {
		return false
	}

	if seen[t] {
		return false
	}

	seen[t] = true

	k := t.Kind()
	if k == reflect.Pointer || k == reflect.Slice || k == reflect.Array {
		return hasExtrasInTypeSeen(t.Elem(), seen)
	}

	if k == reflect.Map {
		return hasExtrasInTypeSeen(t.Key(), seen) || hasExtrasInTypeSeen(t.Elem(), seen)
	}

	if k == reflect.Struct {
		if _, ok := extrasField(t); ok {
			return true
		}

		return structHasExtras(t, seen)
	}

	return false
}

func structHasExtras(t reflect.Type, seen map[reflect.Type]bool) bool {
	for f := range t.Fields() {
		if !f.IsExported() {
			continue
		}

		if hasExtrasInTypeSeen(f.Type, seen) {
			return true
		}
	}

	return false
}

func matchStructField(t reflect.Type, key keySegment, match func(*reflect.StructField, keySegment) bool) (*reflect.StructField, bool) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, false
	}

	n := t.NumField()
	for i := range n {
		f := fieldAt(t, i)
		if !match(f, key) {
			continue
		}

		return f, true
	}

	return nil, false
}

func derefTargetType(v reflect.Value) reflect.Type {
	t := v.Type()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}
