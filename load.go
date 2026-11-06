package goconf

import (
	"errors"
	"reflect"
)

// errLoadAnyCast is the impossible-by-construction cast failure path.
var errLoadAnyCast = errors.New("decode target does not hold the requested type")

// Load discovers, merges, and decodes one config into T, strictly: the
// caller declares the config NAME and the required FORMAT; the details
// are in docs/load.md.
func Load[T any](spec Spec, opts ...Option) (T, []Provenance, error) {
	var zero T
	if err := spec.Format.check(); err != nil {
		return zero, nil, err
	}

	s := settings{search: nil, globalDir: "", defaults: nil}
	for _, o := range opts {
		o.apply(&s)
	}

	if len(s.search) == 0 {
		s.search = defaultSearchPaths(s, spec.Name)
	}

	files, err := discover(Discovery{Search: s.search, Name: spec.Name, Format: spec.Format})
	if err != nil {
		return zero, nil, err
	}

	dst := Target{Value: reflect.ValueOf(new(T)).Elem()}

	prov, err := decodeInto(dst, files, s.defaults)
	if err != nil {
		return zero, nil, err
	}

	got, ok := reflect.TypeAssert[T](dst.Value)
	if !ok {
		return zero, nil, errLoadAnyCast
	}

	return got, prov, nil
}

// decodeInto reads every file, layers their presence, decodes in order
// (later sources override earlier ones), then fills absents from the
// defaults; each source decodes through its own file's format.
func decodeInto(dst Target, files []FoundFile, defaults []LayeredDefault) ([]Provenance, error) {
	sources, prov, err := readSources(files)
	if err != nil {
		return nil, err
	}

	present, err := unionPresence(sources, fileFormats(files))
	if err != nil {
		return nil, err
	}

	for i, src := range sources {
		if err := decodeFormat(src, dst, files[i].Format); err != nil {
			return nil, err
		}
	}

	if err := fillAbsents(dst, present, defaults); err != nil {
		return nil, err
	}

	return prov, nil
}

// fileFormats renders one format per file, in discovery order.
func fileFormats(files []FoundFile) []Format {
	out := make([]Format, 0, len(files))
	for _, f := range files {
		out = append(out, f.Format)
	}

	return out
}
