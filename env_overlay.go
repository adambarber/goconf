package goconf

import (
	"os"
	"reflect"
	"strings"
)

// envRawValue is one environment variable's uninterpreted text.
type envRawValue string

// envBinding couples a variable name to a leaf path. Discovery never holds
// a destination Value: enabling WithEnv must not change the config's shape
// before a variable is read.
type envBinding struct {
	name EnvName
	path FieldPath
}

// envBindings derives leaf paths from the target type. Unrecognized
// variables are ignored; a marked field is a leaf even when it is a struct.
func envBindings(root reflect.Value, prefix EnvPrefix) ([]envBinding, error) {
	var out []envBinding

	seen := make(map[EnvName]FieldPath)

	if err := collectEnvBindings(root.Type(), prefix, "", seen, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func collectEnvBindings(t reflect.Type, prefix EnvPrefix, parent FieldPath, seen map[EnvName]FieldPath, out *[]envBinding) error {
	t = elemType(t)
	if t.Kind() != reflect.Struct {
		return nil
	}

	for i := range t.NumField() {
		f := fieldAt(t, i)
		if !f.IsExported() || string(tagPathOf(f)) == "-" {
			continue
		}

		path := childPath(parent, tagPathOf(f))
		if err := collectEnvField(f, path, prefix, seen, out); err != nil {
			return err
		}
	}

	return nil
}

func collectEnvField(f *reflect.StructField, path FieldPath, prefix EnvPrefix, seen map[EnvName]FieldPath, out *[]envBinding) error {
	if !envIsLeaf(f) {
		return collectEnvBindings(f.Type, prefix, path, seen, out)
	}

	name := envNameFor(prefix, path)
	if other, exists := seen[name]; exists && other != path {
		return envAmbiguousRefusal(name, other, path)
	}

	seen[name] = path
	*out = append(*out, envBinding{name: name, path: path})

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

// envNameFor spells one canonical path as an environment variable name.
func envNameFor(prefix EnvPrefix, path FieldPath) EnvName {
	upper := strings.ToUpper(strings.ReplaceAll(string(path), ".", "_"))

	return EnvName(string(prefix) + "_" + upper)
}

// envLeafValue resolves one binding against the destination only after its
// variable has been read. Pointers along the path are allocated here; no
// allocation happens during discovery.
func envLeafValue(root reflect.Value, path FieldPath) reflect.Value {
	cur := derefAlloc(root)
	for seg := range strings.SplitSeq(string(path), ".") {
		cur = derefAlloc(cur)

		f, ok := fieldMatching(cur.Type(), keySegment(seg))
		if !ok {
			return reflect.Value{}
		}

		cur = derefAlloc(cur.FieldByIndex(f.Index))
	}

	return cur
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
		next, err := applyEnvBinding(dst, binding, present)
		if err != nil {
			return present, err
		}

		present = next
	}

	return present, nil
}

// applyEnvBinding fills one binding when its variable is set. Allocation
// happens here, after LookupEnv, so an unset variable leaves the shape alone.
func applyEnvBinding(dst Target, binding envBinding, present KeySet) (KeySet, error) {
	raw, ok := os.LookupEnv(string(binding.name))
	if !ok {
		return present, nil
	}

	if err := checkOneSecretPath(binding.path, dst.Value.Type(), FilePath(binding.name)); err != nil {
		return present, err
	}

	leaf := envLeafValue(dst.Value, binding.path)
	if !leaf.IsValid() {
		return present, envParseRefusal(binding.name, dst.Value.Type())
	}

	if err := setEnvValue(binding.name, leaf, envRawValue(raw)); err != nil {
		return present, err
	}

	return present.Add(binding.path), nil
}
