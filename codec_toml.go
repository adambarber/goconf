package goconf

import (
	"bytes"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adambarber/gorefuse"
)

// decodeTOMLStrict decodes TOML bytes into the target, refusing unknown
// fields and type mismatches with the offending field named.
func decodeTOMLStrict(data SourceBytes, file FilePath, dst Target) error {
	md, err := toml.NewDecoder(bytes.NewReader(data)).Decode(dst.Value.Addr().Interface())
	if err != nil {
		return tomlRefusal(file, errTOMLValueFree(err))
	}

	return tomlUndecoded(&md, file)
}

// tomlRefusal rewrites one BurntSushi parse refusal as the
// ErrDecoderRefused carrier: the stripped detail survives as the cause.
func tomlRefusal(file FilePath, err error) error {
	return refuseDecoder(
		gorefuse.At(gorefuse.File(file), 0,
			gorefuse.Subject("TOML parse"),
			gorefuse.Why("fix the error the TOML decoder reported.")), err)
}

// tomlUndecoded refuses keys present in the file but missing from the
// target: strictness matches JSON and YAML.
func tomlUndecoded(md *toml.MetaData, file FilePath) error {
	if len(md.Undecoded()) == 0 {
		return nil
	}

	return refuseAt(ErrUnknownField, file, 0,
		gorefuse.Subject(string(joinKeys(md.Undecoded()))),
		gorefuse.Why("the decode target does not declare them; declare the fields in the target or remove them from the file."))
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
func presentTOMLKeys(data SourceBytes, file FilePath, t reflect.Type) (KeySet, error) {
	keys := KeySet{paths: nil}

	var raw tomlKeyMap

	md, err := toml.NewDecoder(bytes.NewReader(data)).Decode(&raw)
	if err != nil {
		return keys, tomlRefusal(file, errTOMLValueFree(err))
	}

	for _, k := range md.Keys() {
		p, err := canonicalTOMLPath(k, file, t)
		if err != nil {
			return keys, err
		}

		keys = keys.Add(p)
	}

	return keys, nil
}

// canonicalTOMLPath maps one document key through the decode target's
// field vocabulary: each segment matches a field's tag or Go name
// case-insensitively, and the canonical (tagPathOf) spelling is recorded.
func canonicalTOMLPath(k toml.Key, file FilePath, t reflect.Type) (FieldPath, error) {
	path := FieldPath("")

	for _, seg := range k {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		f, ok := fieldMatching(t, keySegment(seg))
		if !ok {
			return "", refuseAt(ErrUnknownField, file, 0,
				gorefuse.Subject("\""+k.String()+"\""),
				gorefuse.Why("the decode target does not declare them; declare the fields in the target or remove them from the file."))
		}

		path = childPath(path, tagPathOf(f))
		t = f.Type
	}

	return path, nil
}

// keySegment is one segment of a document key being mapped to a field.
type keySegment string

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
