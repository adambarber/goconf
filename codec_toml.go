package goconf

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

// decodeTOMLStrict decodes TOML bytes into the target, refusing unknown
// fields and type mismatches with the offending field named.
func decodeTOMLStrict(data SourceBytes, dst Target) error {
	md, err := toml.NewDecoder(bytes.NewReader(data)).Decode(dst.Value.Addr().Interface())
	if err != nil {
		return fmt.Errorf("toml: %w", errTOMLValueFree(err))
	}

	return tomlUndecoded(&md)
}

// errTOMLUnknown is the refusal for a key the target does not declare.
var errTOMLUnknown = errors.New("unknown field")

// tomlUndecoded refuses keys present in the file but missing from the
// target: strictness matches JSON and YAML.
func tomlUndecoded(md *toml.MetaData) error {
	if len(md.Undecoded()) == 0 {
		return nil
	}

	return fmt.Errorf("toml: %w: %s", errTOMLUnknown, joinKeys(md.Undecoded()))
}

// keyQuote wraps one key in quotes for a refusal message.
func keyQuote(k toml.Key) keyRender {
	return keyRender(`"` + k.String() + `"`)
}

// keyRender is one rendered key inside a refusal message.
type keyRender string

// joinKeys renders undecoded keys as a comma-separated list.
func joinKeys(keys []toml.Key) keyRender {
	parts := make([]keyRender, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, keyQuote(k))
	}

	return keyRender(strings.Join(toStrings(parts), ", "))
}

// toStrings adapts rendered keys to strings.Join.
func toStrings(parts []keyRender) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, string(p))
	}

	return out
}

// presentTOMLKeys records the dotted field paths of every key present in
// the TOML document: the decoder's key report covers table headers and
// leaf keys, dotted in document order; values are never inspected.
func presentTOMLKeys(data SourceBytes, t reflect.Type) (KeySet, error) {
	keys := KeySet{paths: nil}

	var raw tomlKeyMap

	md, err := toml.NewDecoder(bytes.NewReader(data)).Decode(&raw)
	if err != nil {
		return keys, fmt.Errorf("toml: %w", err)
	}

	for _, k := range md.Keys() {
		p, err := canonicalTOMLPath(k, t)
		if err != nil {
			return keys, err
		}

		keys = keys.Add(p)
	}

	return keys, nil
}

// keySegment is one segment of a document key being mapped to a field.
type keySegment string

// errTOMLKeyUnmappable is the refusal for a document key whose segments
// cannot be mapped to the decode target's field vocabulary.
var errTOMLKeyUnmappable = errors.New("cannot map key to the decode target's fields")

// canonicalTOMLPath maps one document key through the decode target's
// field vocabulary: each segment matches a field's tag or Go name
// case-insensitively, and the canonical (tagPathOf) spelling is recorded.
func canonicalTOMLPath(k toml.Key, t reflect.Type) (FieldPath, error) {
	path := FieldPath("")

	for _, seg := range k {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		f, ok := fieldMatching(t, keySegment(seg))
		if !ok {
			return "", fmt.Errorf("toml: key %q: %w", k.String(), errTOMLKeyUnmappable)
		}

		path = childPath(path, tagPathOf(f))
		t = f.Type
	}

	return path, nil
}

// fieldMatching returns the struct field whose tag or Go name matches
// seg case-insensitively, unexported fields skipped.
func fieldMatching(t reflect.Type, seg keySegment) (*reflect.StructField, bool) {
	for i := range t.NumField() {
		f := fieldAt(t, i)
		if !f.IsExported() {
			continue
		}

		if strings.EqualFold(string(tagPathOf(f)), string(seg)) {
			return f, true
		}
	}

	return nil, false
}

// tomlKeyMap is the decode target of the key walk, keyed by table name.
type tomlKeyMap map[tomlKey]toml.Primitive

// tomlKey is one top-level TOML table name.
type tomlKey string
