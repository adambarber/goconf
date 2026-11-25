package goconf

import (
	"reflect"
	"strings"

	"github.com/adambarber/gorefuse"
)

// defaultValueMap maps a canonical target path to the default value at it.
type defaultValueMap map[FieldPath]reflect.Value

// normalizeDefault accepts a struct or a non-nil pointer to one, returning the
// struct value that fills target fields.
func normalizeDefault(v reflect.Value) (reflect.Value, error) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}, badDefaultRefusal("a nil default pointer")
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, badDefaultRefusal(gorefuse.Subject("a default value of kind " + v.Kind().String()))
	}

	return v, nil
}

// badDefaultRefusal names the option that caused the bad shape.
func badDefaultRefusal(what gorefuse.Subject) error {
	return refuseWhy(ErrBadDefaultShape, gorefuse.Subject("WithDefaults "+string(what)),
		gorefuse.Why("WithDefaults needs a struct or pointer to struct whose fields fill the target."))
}

// defaultValueMapOf records every defaultable field path, recursing through
// schema structs. Opaque leaves contribute only their own path.
func defaultValueMapOf(v reflect.Value) (defaultValueMap, error) {
	out := defaultValueMap{}
	if err := recordDefaultValues(v, "", out); err != nil {
		return nil, err
	}

	return out, nil
}

// eachDefaultField walks the schema fields of one struct, handing the shared
// field/path/value triple to the caller so the walk is written only once.
func eachDefaultField(v reflect.Value, prefix FieldPath, visit func(*reflect.StructField, FieldPath, reflect.Value) error) error {
	t := v.Type()
	for i := range t.NumField() {
		f := fieldAt(t, i)
		if !isDefaultableField(f) {
			continue
		}

		if err := visit(f, childPath(prefix, tagPathOf(f)), v.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

// recordDefaultValues is the recursive half of defaultValueMapOf.
func recordDefaultValues(v reflect.Value, prefix FieldPath, out defaultValueMap) error {
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	return eachDefaultField(v, prefix, func(_ *reflect.StructField, path FieldPath, field reflect.Value) error {
		out[path] = field

		return recordDefaultValues(field, path, out)
	})
}

// isDefaultableField excludes unexported, skipped, and extension fields:
// defaults never write into an extras/rest container or a json:"-" slot.
func isDefaultableField(f *reflect.StructField) bool {
	if !f.IsExported() || string(tagPathOf(f)) == "-" {
		return false
	}

	tag, ok := f.Tag.Lookup("goconf")
	if !ok {
		return true
	}

	name, _, _ := strings.Cut(tag, ",")

	return name != "extras" && name != "rest"
}

// validateDefault checks the whole default shape before a single write:
// every target path must exist and be assignable (or explicitly convertible).
func validateDefault(dst reflect.Value, values defaultValueMap, prefix FieldPath) error {
	return eachDefaultField(dst, prefix, func(_ *reflect.StructField, path FieldPath, field reflect.Value) error {
		def, ok := values[path]
		if !ok {
			return badDefaultRefusal(gorefuse.Subject("is missing field " + string(path)))
		}

		if _, err := convertDefault(def, field.Type(), path); err != nil {
			return err
		}

		return validateNested(field, def, path)
	})
}

// validateNested descends into schema structs on both sides, including
// pointer-to-struct pairs, so nested shape mismatches refuse before writes.
func validateNested(dst, def reflect.Value, path FieldPath) error {
	if !isSchemaStruct(dst) && !isSchemaPointer(dst) {
		return nil
	}

	if !isSchemaStruct(def) && !isSchemaPointer(def) {
		return defaultMismatch(path, def.Type())
	}

	left, right := schemaElem(dst), schemaElem(def)
	if left.Kind() != reflect.Struct || right.Kind() != reflect.Struct {
		return defaultMismatch(path, def.Type())
	}

	values := defaultValueMap{}
	if err := recordDefaultValues(right, path, values); err != nil {
		return err
	}

	return validateDefault(left, values, path)
}
