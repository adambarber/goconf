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
		if err := checkDefaultShape(d.value); err != nil {
			// the option-construction check did not run (a hand-built
			// LayeredDefault): refuse with a domain error, never panic
			return err
		}

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

// copyMissing copies absent leaf fields from the defaults value into the
// target, recursing through struct fields with presence decided per leaf.
func copyMissing(dst, def reflect.Value, prefix FieldPath, present KeySet) error {
	if dst.Kind() != reflect.Struct {
		return nil
	}

	t := dst.Type()
	for i := range t.NumField() {
		if err := copyField(dst.Field(i), def.Field(i), childPath(prefix, tagPathOf(fieldAt(t, i))), present); err != nil {
			return err
		}
	}

	return nil
}

// tagPathOf returns the config key of one struct field: the first
// non-empty of its json, yaml, and toml tags, or the Go name.
func tagPathOf(f *reflect.StructField) FieldPath {
	for _, tagKey := range []string{"json", "yaml", "toml"} {
		tag, ok := f.Tag.Lookup(tagKey)
		if !ok {
			continue
		}

		name, _, _ := strings.Cut(tag, ",")
		if name != "" {
			return FieldPath(name)
		}
	}

	return FieldPath(f.Name)
}

// copyField fills one field: structs recurse so presence is decided at
// each leaf; leaves fill only when the leaf path itself is absent.
func copyField(dst, def reflect.Value, path FieldPath, present KeySet) error {
	if dst.Kind() == reflect.Struct {
		return copyMissing(dst, def, path, present)
	}

	if present.Has(path) || !dst.CanSet() {
		return nil
	}

	dst.Set(def)

	return nil
}

// childPath joins one field to its parent's dotted path.
func childPath(parent, name FieldPath) FieldPath {
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
