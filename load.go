package goconf

import (
	"errors"
	"reflect"

	"github.com/adambarber/gorefuse"
)

// Load discovers, merges, and decodes one config into T, strictly: the
// caller declares the config NAME and optionally a FORMAT; the details
// are in docs/load.md.
func Load[T any](spec Spec, opts ...Option) (T, []Provenance, error) {
	got, prov, _, err := LoadWithMeta[T](spec, opts...)
	return got, prov, err
}

// Redact renders cfg the way fmt.Sprintf("%+v", cfg) would, except that every
// field tagged secret:"true" renders as gorefuse.Redacted. Marked parents
// cover their descendants; an explicit secret:"false" overrides inheritance.
func Redact[T any](cfg T) Rendered {
	return renderValue(reflect.ValueOf(cfg))
}

// LoadWithMeta behaves like Load and also returns per-occurrence metadata.
// On refusal it returns zero config, nil provenance, zero Meta, and the error.
// See docs/load.md for metadata semantics and query helpers.
func LoadWithMeta[T any](spec Spec, opts ...Option) (T, []Provenance, Meta, error) {
	var zero T

	s, err := prepareLoad(spec, opts...)
	if err != nil {
		return zero, nil, Meta{}, err
	}

	if len(s.search) == 0 {
		s.search = defaultSearchPaths(&s, spec.Name)
	}

	files, err := discover(Discovery{Search: s.search, Name: spec.Name, Format: spec.Format})
	if err != nil {
		if !envOrOptionalAbsence(&s, err) {
			return zero, nil, Meta{}, err
		}

		// True absence is allowed when the caller opted in or declared an
		// env overlay: the env values are then the whole config source.

		files = []FoundFile{}

		if s.optional {
			warnOptionalAbsence()
		}
	}

	dst, err := newTarget[T](&s)
	if err != nil {
		return zero, nil, Meta{}, err
	}

	sources, prov, err := readSources(files, s.maxBytes)
	if err != nil {
		return zero, nil, Meta{}, err
	}

	return loadAcquired(zero, dst, &s, acquired{sources: sources, files: files, prov: prov})
}

// envOrOptionalAbsence reports whether a discovery fault is true absence that
// this load may proceed through: the caller opted in, or declared env values.
func envOrOptionalAbsence(s *settings, err error) bool {
	return errors.Is(err, ErrNothingFound) && (s.optional || s.envPrefix != "")
}

// acquired is one load's sources and their provenance, after discovery and
// before decoding. Disk discovery and LoadBytes both produce this same shape.
type acquired struct {
	sources []SourceBytes
	files   []FoundFile
	prov    []Provenance
}

// newTarget builds the decode target after checking defaults against it, so a
// bad default refuses before any source bytes are read.
func newTarget[T any](s *settings) (Target, error) {
	dst := Target{Value: reflect.ValueOf(new(T)).Elem()}
	if err := validateDefaults(dst.Value, s.defaults); err != nil {
		return Target{}, err
	}

	return dst, nil
}

// loadAcquired decodes and casts one acquired load; every entry point funnels
// through here so the strict contract has one owner.
func loadAcquired[T any](zero T, dst Target, s *settings, src acquired) (T, []Provenance, Meta, error) {
	for i, data := range src.sources {
		if err := checkByteCap(len(data), src.files[i].Path, s.maxBytes); err != nil {
			return zero, nil, Meta{}, err
		}
	}

	meta, err := decodeSources(dst, src.sources, src.files, s)
	if err != nil {
		return zero, nil, Meta{}, err
	}

	got, err := finishLoad[T](zero, dst)
	if err != nil {
		return zero, nil, Meta{}, err
	}

	return got, src.prov, meta, nil
}

// finishLoad casts the decoded target to T, refusing a construction fault.
func finishLoad[T any](zero T, dst Target) (T, error) {
	got, ok := reflect.TypeAssert[T](dst.Value)
	if !ok {
		return zero, refuseWhy(ErrTargetCast,
			gorefuse.Subject("the decode target"),
			gorefuse.Why("the decode target does not hold the requested type; this is a construction fault."))
	}

	return got, nil
}

// prepareLoad validates the spec and folds the options into one settings
// value; disk and in-memory entry points share this boundary.
func prepareLoad(spec Spec, opts ...Option) (settings, error) {
	if err := spec.Format.check(); err != nil {
		return settings{}, err
	}

	var s settings

	s.maxBytes = defaultByteCap
	for _, o := range opts {
		o.apply(&s)
	}

	return s, nil
}

// decodeSources runs the strict decode, env overlay, default fill, and
// metadata pass over already-acquired bytes; both entry points share it.
func decodeSources(dst Target, sources []SourceBytes, files []FoundFile, s *settings) (Meta, error) {
	present, meta, err := collectPresenceAndMeta(sources, files, dst.Value.Type())
	if err != nil {
		return Meta{}, err
	}

	for i, src := range sources {
		if decodeErr := decodeFormat(src, files[i].Path, dst, files[i].Format); decodeErr != nil {
			return Meta{}, decodeErr
		}
	}

	envKeys, err := applyEnvOverlay(dst, s.envPrefix)
	if err != nil {
		return Meta{}, err
	}

	for _, path := range envKeys.pathsList() {
		present = present.Add(path)
	}

	if err := fillAbsents(dst, present, s.defaults); err != nil {
		return Meta{}, err
	}

	enrichMeta(&meta, dst)

	return meta, nil
}
