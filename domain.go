package goconf

import "reflect"

// FieldPath names one config field as it appears in a file: the dotted path
// from the document root, for example "server.tls.cert_file".
type FieldPath string

// SourceBytes holds the raw bytes of one config source.
type SourceBytes []byte

// Extension is a file extension including the leading dot: ".json".
type Extension string

// FileName is the name of a config file on disk, including its extension.
type FileName string

// DirPath is the path of a directory searched for config files.
type DirPath string

// FilePath is the path of a file on disk.
type FilePath string

// Target is the reflect view of the destination a format decodes into.
// Value must be addressable and is always set by the caller through
// reflect.ValueOf(&v).Elem(); the library never takes an untyped parameter.
type Target struct {
	Value reflect.Value
}

// pathSet is the storage behind a KeySet.
type pathSet map[FieldPath]struct{}

// KeySet is the set of field paths present in one source document.
type KeySet struct {
	paths pathSet
}

// Has reports whether the field path is present in the source document.
func (k KeySet) Has(p FieldPath) bool {
	_, ok := k.paths[p]
	return ok
}

// Add records one present field path.
func (k KeySet) Add(p FieldPath) KeySet {
	if k.paths == nil {
		k.paths = make(pathSet)
	}

	k.paths[p] = struct{}{}

	return k
}

// pathsList lists the present field paths in map order.
func (k KeySet) pathsList() []FieldPath {
	out := make([]FieldPath, 0, len(k.paths))
	for p := range k.paths {
		out = append(out, p)
	}

	return out
}
