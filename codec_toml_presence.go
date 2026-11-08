package goconf

import (
	"bytes"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adambarber/gorefuse"
)

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
				gorefuse.Subject(`"`+k.String()+`"`),
				gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
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
