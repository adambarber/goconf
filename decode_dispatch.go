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
	switch f {
	case FormatJSON:
		return presentJSONKeys(src, file)
	case FormatYAML:
		return presentYAMLKeys(src, file)
	case FormatTOML:
		return presentTOMLKeys(src, file, t)
	case undetermined:
		return KeySet{}, refuseWhy(ErrUnknownFormat,
			gorefuse.Subject("the undetermined format reached the presence walk"),
			gorefuse.Why("discovery stamps each file's format; this is a construction fault."))
	default:
		return KeySet{}, refuseWhy(ErrUnknownFormat,
			gorefuse.Subject("a format no constant spells"),
			gorefuse.Why("spell a format with FormatJSON, FormatYAML, or FormatTOML."))
	}
}

// unionPresence merges the presence views of all sources, later sources
// adding to earlier ones. Sources may carry different formats: each is
// decoded through its own.
func unionPresence(sources []SourceBytes, files []FoundFile, t reflect.Type) (KeySet, error) {
	union := KeySet{paths: nil}

	for i, src := range sources {
		keys, err := presenceOf(src, files[i].Format, files[i].Path, t)
		if err != nil {
			return union, err
		}

		for _, p := range keys.pathsList() {
			union = union.Add(p)
		}
	}

	return union, nil
}
