package goconf

import (
	"reflect"

	"github.com/BurntSushi/toml"
)

func valueTypeOf() reflect.Type { return reflect.TypeFor[Value]() }

// tomlShadowType swaps Value fields for toml.Primitive so the single parse
// captures both shapes. Opaque leaves like time.Time keep their own type.
func tomlShadowType(t reflect.Type) reflect.Type {
	if t == valueTypeOf() {
		return reflect.TypeFor[toml.Primitive]()
	}

	if !hasValueType(t, map[reflect.Type]bool{}) {
		return t
	}

	if t.Kind() == reflect.Struct {
		return shadowStructType(t)
	}

	return shadowCompositeType(t)
}

func shadowCompositeType(t reflect.Type) reflect.Type {
	switch t.Kind() {
	case reflect.Map:
		return reflect.MapOf(t.Key(), shadowElem(t))
	case reflect.Slice:
		return reflect.SliceOf(shadowElem(t))
	case reflect.Array:
		return reflect.ArrayOf(t.Len(), shadowElem(t))
	case reflect.Pointer:
		return reflect.PointerTo(shadowElem(t))
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.String, reflect.Struct, reflect.UnsafePointer:
		return t
	}

	return t
}

func shadowElem(t reflect.Type) reflect.Type {
	return tomlShadowType(t.Elem())
}

func shadowStructType(t reflect.Type) reflect.Type {
	fields := make([]reflect.StructField, 0, t.NumField())

	for f := range t.Fields() {
		f.Type = tomlShadowType(f.Type)
		fields = append(fields, f)
	}

	return reflect.StructOf(fields)
}

// hasValueType reports whether t contains a Value anywhere in its shape.
func hasValueType(t reflect.Type, seen map[reflect.Type]bool) bool {
	if seen[t] {
		return false
	}

	seen[t] = true

	if t == valueTypeOf() {
		return true
	}

	if t.Kind() == reflect.Struct {
		return structHasValueType(t, seen)
	}

	if isCompositeType(t) {
		return hasValueType(t.Elem(), seen)
	}

	return false
}

func structHasValueType(t reflect.Type, seen map[reflect.Type]bool) bool {
	for f := range t.Fields() {
		if hasValueType(f.Type, seen) {
			return true
		}
	}

	return false
}

func isCompositeType(t reflect.Type) bool {
	k := t.Kind()

	return k == reflect.Map || k == reflect.Slice || k == reflect.Array || k == reflect.Pointer
}
