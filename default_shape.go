package goconf

import (
	"encoding"
	"encoding/json"
	"reflect"

	"github.com/adambarber/gorefuse"
)

// schemaElem returns the struct behind a schema pointer, constructing a zero
// value for a nil target pointer so validation can still see its fields.
func schemaElem(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Pointer {
		return v
	}

	return reflect.New(v.Type().Elem()).Elem()
}

// isSchemaPointer reports whether v points at a struct with defaultable fields.
func isSchemaPointer(v reflect.Value) bool {
	return v.Kind() == reflect.Pointer && isSchemaStruct(schemaElem(v))
}

// isSchemaStruct reports whether v is a struct whose fields participate in
// the config schema. time.Time and similar opaque leaves do not.
func isSchemaStruct(v reflect.Value) bool {
	if v.Kind() != reflect.Struct {
		return false
	}

	t := v.Type()
	if implementsUnmarshaler(t) {
		return false
	}

	for i := range t.NumField() {
		if isDefaultableField(fieldAt(t, i)) {
			return true
		}
	}

	return false
}

// implementsUnmarshaler treats a struct that decodes itself as an opaque
// leaf, even when its Go fields are exported.
func implementsUnmarshaler(t reflect.Type) bool {
	return t.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) ||
		reflect.PointerTo(t).Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) ||
		t.Implements(reflect.TypeFor[json.Unmarshaler]()) ||
		reflect.PointerTo(t).Implements(reflect.TypeFor[json.Unmarshaler]())
}

// convertDefault returns def as want, refusing a type that cannot fill it.
func convertDefault(def reflect.Value, want reflect.Type, path FieldPath) (reflect.Value, error) {
	if !def.IsValid() {
		return reflect.Value{}, defaultMismatch(path, nil)
	}

	if def.Type().AssignableTo(want) {
		return def, nil
	}

	if def.Type().ConvertibleTo(want) {
		return def.Convert(want), nil
	}

	return reflect.Value{}, defaultMismatch(path, def.Type())
}

// defaultMismatch refuses a default whose type cannot fill a target field.
func defaultMismatch(path FieldPath, got reflect.Type) error {
	subject := gorefuse.Subject("field " + string(path))
	if got != nil {
		subject = gorefuse.Subject(string(subject) + " of type " + got.String())
	}

	return refuseWhy(ErrBadDefaultShape, subject,
		gorefuse.Why("the default cannot fill this field; give it the field's type."))
}
