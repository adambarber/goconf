package goconf

import (
	"os"
	"reflect"
	"strings"
)

// envRawValue is one environment variable's uninterpreted text.
type envRawValue string

// envBinding couples one environment variable name to its schema leaf.
type envBinding struct {
	name  EnvName
	path  FieldPath
	value reflect.Value
}

// envBindings derives every addressable leaf from the decode target. Only
// schema fields are considered: an unrecognized variable is ignored rather
// than guessed at. A marked field is a leaf even when it is a struct.
func envBindings(root reflect.Value, prefix EnvPrefix) ([]envBinding, error) {
	var out []envBinding

	seen := make(map[EnvName]FieldPath)

	if err := collectEnvBindings(root, prefix, "", seen, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func collectEnvBindings(v reflect.Value, prefix EnvPrefix, parent FieldPath, seen map[EnvName]FieldPath, out *[]envBinding) error {
	t := elemType(v.Type())

	for i := range t.NumField() {
		f := fieldAt(t, i)
		if !f.IsExported() || string(tagPathOf(f)) == "-" {
			continue
		}

		path := childPath(parent, tagPathOf(f))

		if err := collectEnvField(v, i, f, path, prefix, seen, out); err != nil {
			return err
		}
	}

	return nil
}

func collectEnvField(v reflect.Value, i int, f *reflect.StructField, path FieldPath, prefix EnvPrefix, seen map[EnvName]FieldPath, out *[]envBinding) error {
	if !envIsLeaf(f) {
		return collectEnvBindings(envStructField(v, i), prefix, path, seen, out)
	}

	name := envNameFor(prefix, path)
	if other, exists := seen[name]; exists && other != path {
		return envAmbiguousRefusal(name, other, path)
	}

	seen[name] = path

	*out = append(*out, envBinding{name: name, path: path, value: v.Field(i)})

	return nil
}

// envIsLeaf reports whether one field is filled as a whole from an env value.
func envIsLeaf(f *reflect.StructField) bool {
	t := elemType(f.Type)
	if declaredSecret(f) == secretMarked || implementsUnmarshaler(t) {
		return true
	}

	return t.Kind() != reflect.Struct
}

// envStructField returns a field's value with pointers dereferenced for descent.
func envStructField(v reflect.Value, i int) reflect.Value {
	return derefAlloc(v.Field(i))
}

// envNameFor spells one canonical path as an environment variable name.
func envNameFor(prefix EnvPrefix, path FieldPath) EnvName {
	upper := strings.ToUpper(strings.ReplaceAll(string(path), ".", "_"))

	return EnvName(string(prefix) + "_" + upper)
}

// applyEnvOverlay overlays process variables onto an already-decoded config.
// It returns the paths that env filled, so defaults cannot overwrite them.
func applyEnvOverlay(dst Target, prefix EnvPrefix) (KeySet, error) {
	present := KeySet{paths: nil}
	if prefix == "" {
		return present, nil
	}

	bindings, err := envBindings(dst.Value, prefix)
	if err != nil {
		return present, err
	}

	for _, binding := range bindings {
		raw, ok := os.LookupEnv(string(binding.name))
		if !ok {
			continue
		}

		if err := checkOneSecretPath(binding.path, dst.Value.Type(), FilePath(binding.name)); err != nil {
			return present, err
		}

		if err := parseEnvBinding(binding, envRawValue(raw)); err != nil {
			return present, err
		}

		present = present.Add(binding.path)
	}

	return present, nil
}
