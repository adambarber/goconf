package goconf

import "github.com/adambarber/gorefuse"

// LoadBytes decodes in-memory bytes into T under the same strict rules as
// Load. spec.Name is provenance identity only; no file is discovered or read.
func LoadBytes[T any](data []byte, spec Spec, opts ...Option) (T, []Provenance, error) {
	got, prov, _, err := LoadBytesWithMeta[T](data, spec, opts...)
	return got, prov, err
}

// LoadBytesWithMeta behaves like LoadBytes and also returns per-occurrence
// metadata. On refusal it returns zero config, nil provenance, zero Meta,
// and the error.
func LoadBytesWithMeta[T any](data []byte, spec Spec, opts ...Option) (T, []Provenance, Meta, error) {
	var zero T

	return withLoadSettings(zero, spec, func(zero T, s settings) (T, []Provenance, Meta, error) {
		if spec.Format == undetermined {
			return zero, nil, Meta{}, refuseWhy(ErrUnknownFormat,
				gorefuse.Subject("the in-memory source has no declared format"),
				gorefuse.Why("declare a format on Spec; accepted: "+acceptedFormats+"."))
		}

		dst, err := newTarget[T](&s)
		if err != nil {
			return zero, nil, Meta{}, err
		}

		file := FoundFile{Path: FilePath(string(spec.Name)), Format: spec.Format, Leg: DirPath("")}
		sources := []SourceBytes{SourceBytes(data)}
		prov := []Provenance{provenanceOf(file, sources[0])}

		return loadAcquired(zero, dst, &s, acquired{sources: sources, files: []FoundFile{file}, prov: prov})
	}, opts...)
}
