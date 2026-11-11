package goconf

import (
	"reflect"

	"github.com/adambarber/gorefuse"
)

// decodeFormat decodes one source's bytes into the target through the
// format's strict decoder; the undetermined arm is unreachable by
// construction (discovery stamps each file's format; Load checks spec's).
func decodeFormat(src SourceBytes, file FilePath, dst Target, f Format) error {
	switch f {
	case FormatJSON:
		return decodeJSONStrict(src, file, dst)
	case FormatYAML:
		return decodeYAMLStrict(src, file, dst)
	case FormatTOML:
		return decodeTOMLStrict(src, file, dst)
	case undetermined:
		return refuseWhy(ErrUnknownFormat,
			gorefuse.Subject("the undetermined format reached the decode path"),
			gorefuse.Why("discovery stamps each file's format; this is a construction fault."))
	default:
		return refuseWhy(ErrUnknownFormat,
			gorefuse.Subject("a format no constant spells"),
			gorefuse.Why("spell a format with FormatJSON, FormatYAML, or FormatTOML."))
	}
}

// presenceOf records the field paths present in one source, per format.
// TOML presence is mapped through the decode target's field vocabulary,
// so t is the decode target's type.
func presenceOf(src SourceBytes, f Format, file FilePath, t reflect.Type) (KeySet, error) {
	keys, _, err := collectSourceMeta(src, f, file, 0, t)
	return keys, err
}

// collectPresenceAndMeta merges presence and collects per-occurrence metadata
// for all sources in discovery/load order.
func collectPresenceAndMeta(sources []SourceBytes, files []FoundFile, t reflect.Type) (KeySet, Meta, error) {
	union := KeySet{paths: nil}
	meta := Meta{entries: nil}

	for i, src := range sources {
		keys, entries, err := collectSourceMeta(src, files[i].Format, files[i].Path, i, t)
		if err != nil {
			return union, Meta{}, err
		}

		for _, p := range keys.pathsList() {
			union = union.Add(p)
		}

		meta.entries = append(meta.entries, entries...)
	}

	return union, meta, nil
}

// collectSourceMeta collects presence and per-occurrence metadata for one
// source through its format's parser/walk.
func collectSourceMeta(src SourceBytes, f Format, file FilePath, sourceIndex int, t reflect.Type) (KeySet, []Occurrence, error) {
	switch f {
	case FormatJSON:
		return collectJSONMeta(src, file, sourceIndex)
	case FormatYAML:
		return collectYAMLMeta(src, file, sourceIndex)
	case FormatTOML:
		return collectTOMLMeta(src, file, sourceIndex, t)
	case undetermined:
		return KeySet{}, nil, refuseWhy(ErrUnknownFormat,
			gorefuse.Subject("the undetermined format reached the presence walk"),
			gorefuse.Why("discovery stamps each file's format; this is a construction fault."))
	default:
		return KeySet{}, nil, refuseWhy(ErrUnknownFormat,
			gorefuse.Subject("a format no constant spells"),
			gorefuse.Why("spell a format with FormatJSON, FormatYAML, or FormatTOML."))
	}
}

// enrichMeta resolves effective values for recorded occurrences from the
// final decoded configuration after layering and defaults, and marks the
// effective (winning) occurrence per indexed path in load order.
func enrichMeta(meta *Meta, dst Target) {
	if meta == nil {
		return
	}

	last := map[OccurrencePath]int{}
	for i := range meta.entries {
		last[meta.entries[i].Path] = i
	}

	for i := range meta.entries {
		e := &meta.entries[i]
		if last[e.Path] != i {
			e.Effective = false

			continue
		}

		if v, ok := lookupEffectiveValue(dst.Value, e.Path); ok {
			e.Value = v
			e.Effective = true

			continue
		}

		// Path not resolvable in effective config (e.g. superseded container
		// shape differs); keep source-captured value and mark non-effective.
		e.Effective = false
	}
}
