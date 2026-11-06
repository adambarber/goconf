package goconf

import (
	"reflect"
	"strings"
)

// fillAbsents applies the defaults layers to the fields no source set:
// a field present in any source (including an explicit null) is kept as
// decoded; only true absences are filled, layer by layer.
func fillAbsents(dst Target, present KeySet, defaults []LayeredDefault) error {
	for _, d := range defaults {
		if err := fillOne(dst, present, d); err != nil {
			return err
		}

		present = absorbPresent(present, d.value)
	}

	return nil
}

// fillOne copies each absent field of one defaults layer into the target.
func fillOne(dst Target, present KeySet, d LayeredDefault) error {
	return copyMissing(dst.Value, d.value, "", present)
}

// absorbPresent extends the presence view with the fields a defaults
// layer has now filled, so later layers do not re-fill them.
func absorbPresent(present KeySet, d reflect.Value) KeySet {
	out := present
	for _, p := range fieldsOf(d, "") {
		out = out.Add(p)
	}

	return out
}

// fieldsOf lists the field paths a defaults layer would fill.
func fieldsOf(v reflect.Value, prefix FieldPath) []FieldPath {
	out := []FieldPath{}

	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := range t.NumField() {
			p := childPath(prefix, tagPathOf(fieldAt(t, i)))
			out = append(out, p)
			out = append(out, fieldsOf(v.Field(i), p)...)
		}
	}
	// every other kind contributes only itself, already recorded by the caller

	return out
}

// copyMissing copies absent fields from the defaults value into the
// target, recursing into nested structs and skipping present leaves.
func copyMissing(dst reflect.Value, def reflect.Value, prefix FieldPath, present KeySet) error {
	if dst.Kind() != reflect.Struct {
		return nil
	}

	t := dst.Type()
	for i := range t.NumField() {
		p := childPath(prefix, tagPathOf(fieldAt(t, i)))
		if present.Has(p) {
			continue
		}

		if err := copyField(dst.Field(i), def.Field(i), p, present); err != nil {
			return err
		}
	}

	return nil
}

// tagPathOf returns the config key of one struct field: the first part
// of its json tag, or the Go field name when untagged.
func tagPathOf(f *reflect.StructField) FieldPath {
	tag, ok := f.Tag.Lookup("json")
	if !ok {
		return FieldPath(f.Name)
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return FieldPath(f.Name)
	}

	return FieldPath(name)
}

// copyField fills one field: leaves are set directly, structs recurse so
// partially-set nested objects keep their source values.
func copyField(dst, def reflect.Value, path FieldPath, present KeySet) error {
	if dst.Kind() == reflect.Struct {
		return copyMissing(dst, def, path, present)
	}

	if !dst.CanSet() {
		return nil
	}

	dst.Set(def)

	return nil
}

// childPath joins one field to its parent's dotted path.
func childPath(parent FieldPath, name FieldPath) FieldPath {
	if parent == "" {
		return name
	}

	return parent + "." + name
}

// fieldAt returns a pointer to one field's metadata, so heavy structs
// pass by pointer.
func fieldAt(t reflect.Type, i int) *reflect.StructField {
	f := t.Field(i)
	return &f
}
