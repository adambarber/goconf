package goconf

import (
	"reflect"

	"github.com/adambarber/gorefuse"
)

// fillAbsents applies each defaults layer only to fields absent from every
// source. A layer is matched by canonical path against the decode target; a
// shape or type that cannot fill the target is refused before any write.
func fillAbsents(dst Target, present KeySet, defaults []layeredDefault) error {
	if len(defaults) == 0 {
		return nil
	}

	root, err := writableStruct(dst.Value)
	if err != nil {
		return err
	}

	for _, d := range defaults {
		values, err := checkedDefaults(root, d.value)
		if err != nil {
			return err
		}

		if err := fillFromValues(root, present, values, ""); err != nil {
			return err
		}

		present = absorbPresent(present, values)
	}

	return nil
}

// validateDefaults checks every layer against the target type before any
// source is read, so a bad default refuses without touching config files.
func validateDefaults(dst reflect.Value, defaults []layeredDefault) error {
	view, err := defaultTargetView(dst)
	if err != nil {
		return err
	}

	for _, d := range defaults {
		if _, err := checkedDefaults(view, d.value); err != nil {
			return err
		}
	}

	return nil
}

// defaultTargetView is a non-mutating view of a (possibly pointer) target.
func defaultTargetView(v reflect.Value) (reflect.Value, error) {
	for v.Kind() == reflect.Pointer {
		v = reflect.New(v.Type().Elem()).Elem()
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, badDefaultRefusal(gorefuse.Subject("a default target of kind " + v.Kind().String()))
	}

	return v, nil
}

// writableStruct returns the struct a pointer target points at, allocating it
// only when defaults actually need to write into it.
func writableStruct(v reflect.Value) (reflect.Value, error) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			if !v.CanSet() {
				return reflect.Value{}, badDefaultRefusal("an unsettable pointer target")
			}

			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, badDefaultRefusal(gorefuse.Subject("a default target of kind " + v.Kind().String()))
	}

	return v, nil
}

// checkedDefaults normalizes one layer and checks its whole shape against the
// target before any field is written.
func checkedDefaults(dst, raw reflect.Value) (defaultValueMap, error) {
	def, err := normalizeDefault(raw)
	if err != nil {
		return nil, err
	}

	values, err := defaultValueMapOf(def)
	if err != nil {
		return nil, err
	}

	if err := validateDefault(dst, values, ""); err != nil {
		return nil, err
	}

	return values, nil
}

// fillFromValues fills each absent field from the default value at its path.
func fillFromValues(dst reflect.Value, present KeySet, values defaultValueMap, prefix FieldPath) error {
	return eachDefaultField(dst, prefix, func(_ *reflect.StructField, path FieldPath, field reflect.Value) error {
		// validateDefault already proved every target path has a value.
		return fillValue(field, values[path], path, present, values)
	})
}

// fillValue handles a pointer field by allocating a fresh element so no
// default backing object is shared with the target.
func fillValue(dst, def reflect.Value, path FieldPath, present KeySet, values defaultValueMap) error {
	if dst.Kind() != reflect.Pointer {
		return fillLeafOrStruct(dst, def, path, present, values)
	}

	if def.Kind() != reflect.Pointer || def.IsNil() {
		return defaultMismatch(path, def.Type())
	}

	if !dst.CanSet() {
		return nil
	}

	elem := reflect.New(dst.Type().Elem())
	if err := fillLeafOrStruct(elem.Elem(), def.Elem(), path, present, values); err != nil {
		return err
	}

	dst.Set(elem)

	return nil
}

// fillLeafOrStruct recurses through schema structs; opaque leaves such as
// time.Time are assigned wholesale after a type check.
func fillLeafOrStruct(dst, def reflect.Value, path FieldPath, present KeySet, values defaultValueMap) error {
	if isSchemaStruct(dst) && isSchemaStruct(def) {
		return fillFromValues(dst, present, values, path)
	}

	if present.Has(path) || !dst.CanSet() {
		return nil
	}

	converted, err := convertDefault(def, dst.Type(), path)
	if err != nil {
		return err
	}

	dst.Set(cloneValue(converted))

	return nil
}

// absorbPresent records the paths a layer filled so later layers stand down.
func absorbPresent(present KeySet, values defaultValueMap) KeySet {
	out := present
	for p := range values {
		out = out.Add(p)
	}

	return out
}
