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

	prov, err := decodeInto(dst, files, spec.Format, s.defaults)
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
// (later sources override earlier ones), then fills absent fields from
// the defaults layers.
func decodeInto(dst Target, files []FoundFile, format Format, defaults []LayeredDefault) ([]Provenance, error) {
	sources, prov, err := readSources(files)
	if err != nil {
		return nil, err
	}

	present, err := unionPresence(sources, format)
	if err != nil {
		return nil, err
	}

	for _, src := range sources {
		if err := decodeFormat(src, dst, format); err != nil {
			return nil, err
		}
	}

	if err := fillAbsents(dst, present, defaults); err != nil {
		return nil, err
	}

	return prov, nil
}
