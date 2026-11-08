package goconf

import (
	"reflect"
	"strings"
	"testing"
)

// innerTagged is the nested struct shared by the tag-combination
// targets: one leaf field whose tags the combination's parent names.
type innerTagged struct {
	V leafValue `json:"v" toml:"v" yaml:"v"`
}

// innerUntagged is the child for the untagged combination: no tags at
// any level, so every walk falls back to Go field names.
type innerUntagged struct {
	V leafValue
}

// leafValue is one config leaf value in the tag-combination fixtures.
type leafValue string

// explicitValue is the value a source sets; defaultValue is what a
// default would fill, so an override is visible.
const (
	explicitValue leafValue = "explicit-value"
	defaultValue  leafValue = "DEFAULT"
)

// tagComboTarget is one parent struct shape per tag mix: the fresh
// decode target, the key vocabulary the walks must agree on, the file
// bytes naming that key per format, and the formats the shape binds.
type tagComboTarget struct {
	key     FieldPath
	shape   reflect.Value
	data    map[Format]SourceBytes
	formats []Format
}

// tagCombos lists every tag mix: json-only, yaml-only, toml-only, all
// three, and untagged (Go-name fallback).
func tagCombos() []tagComboTarget {
	type jsonOnly struct {
		In innerTagged `json:"in"`
	}

	type yamlOnly struct {
		In innerTagged `yaml:"in"`
	}

	type tomlOnly struct {
		In innerTagged `toml:"in"`
	}

	type allTags struct {
		In innerTagged `json:"in" toml:"in" yaml:"in"`
	}

	type untagged struct {
		In innerUntagged
	}

	return []tagComboTarget{
		newCombo(reflect.TypeOf(jsonOnly{In: innerTagged{V: ""}}), "in.v", FormatJSON),
		newCombo(reflect.TypeOf(yamlOnly{In: innerTagged{V: ""}}), "in.v", FormatYAML),
		newCombo(reflect.TypeOf(tomlOnly{In: innerTagged{V: ""}}), "in.v", FormatTOML),
		newCombo(reflect.TypeOf(allTags{In: innerTagged{V: ""}}), "in.v", FormatJSON, FormatYAML, FormatTOML),
		newCombo(reflect.TypeOf(untagged{In: innerUntagged{V: ""}}), "In.V", FormatJSON),
	}
}

// newCombo builds one combination: a fresh decode target of the shape's
// type, the file bytes naming key in each bound format, and the formats.
func newCombo(shape reflect.Type, key FieldPath, formats ...Format) tagComboTarget {
	return tagComboTarget{
		key:     key,
		shape:   reflect.New(shape).Elem(),
		data:    tagData(key, explicitValue),
		formats: formats,
	}
}

// tagData writes one explicit leaf value under the parent of key, per
// format; the leaf is key's last segment.
func tagData(key FieldPath, value leafValue) map[Format]SourceBytes {
	parent, leaf, _ := strings.Cut(string(key), ".")

	return map[Format]SourceBytes{
		FormatJSON: SourceBytes(`{"` + parent + `":{"` + leaf + `":"` + string(value) + `"}}`),
		FormatYAML: SourceBytes(parent + ":\n  " + leaf + ": " + string(value) + "\n"),
		FormatTOML: SourceBytes("[" + parent + "]\n" + leaf + " = \"" + string(value) + "\"\n"),
	}
}

// TestDefaultsMatchPresenceAcrossTagCombinations pins the presence-path
// vocabulary: tagPathOf reads json, yaml, and toml tags, so a default
// never overrides an explicit value, whatever tag mix the target declares.
func TestDefaultsMatchPresenceAcrossTagCombinations(t *testing.T) {
	for _, combo := range tagCombos() {
		for _, format := range combo.formats {
			t.Run(string(combo.key)+" "+string(format), func(t *testing.T) {
				assertDefaultDoesNotOverride(t, combo, format)
			})
		}
	}
}

// assertDefaultDoesNotOverride loads one format's explicit value with a
// default present, and asserts the explicit value survived.
func assertDefaultDoesNotOverride(t *testing.T, combo tagComboTarget, format Format) {
	t.Helper()

	data := combo.data[format]
	dst := combo.shape

	present := presenceOK(t, format, data, dst.Type())
	if !present.Has(combo.key) {
		t.Fatalf("presence walk missed %q: %v", combo.key, present)
	}

	// decode into the fresh target, then fill absents from a default
	// whose leaf differs from the explicit value

	if err := decodeFormat(data, FilePath(string(fileNameJSON)), Target{Value: dst}, format); err != nil {
		t.Fatalf("decode: %v", err)
	}

	def := reflect.New(dst.Type()).Elem()
	def.Field(0).Field(0).SetString(string(defaultValue))

	if err := fillAbsents(Target{Value: dst}, present, []layeredDefault{{value: def}}); err != nil {
		t.Fatalf("fill: %v", err)
	}

	got := dst.Field(0).Field(0).String()
	if got != string(explicitValue) {
		t.Fatalf("default overrode an explicit value: got %q", got)
	}
}

// presenceOK runs one format's presence walk against the combo's target
// shape, failing the test on error.
func presenceOK(t *testing.T, format Format, data SourceBytes, shape reflect.Type) KeySet {
	t.Helper()

	keys, err := presenceOf(data, format, FilePath(string(fileNameJSON)), shape)
	if err != nil {
		t.Fatalf("presence: %v", err)
	}

	return keys
}
