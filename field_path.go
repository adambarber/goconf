package goconf

import (
	"reflect"
	"strings"
)

// declaredTagNames lists the non-empty json, yaml, and toml names of f.
func declaredTagNames(f *reflect.StructField) []string {
	names := []string{}

	for _, tagKey := range []string{"json", "yaml", "toml"} {
		tag, ok := f.Tag.Lookup(tagKey)
		if !ok {
			continue
		}

		name, _, _ := strings.Cut(tag, ",")
		if name != "" {
			names = append(names, name)
		}
	}

	return names
}

// tagPathOf returns the config key of one struct field: the first declared
// tag name, or the Go name when no tag declares one.
func tagPathOf(f *reflect.StructField) FieldPath {
	names := declaredTagNames(f)
	if len(names) > 0 {
		return FieldPath(names[0])
	}

	return FieldPath(f.Name)
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
