package goconf

import (
	"errors"
	"fmt"
)

// Format is one config file format, spelled only by its full extension.
// The zero value is the undetermined format: discovery then demands that
// each directory hold exactly one candidate extension.
type Format string

// The three supported formats, spelled only by their full, unambiguous
// extensions; ".yml" is deliberately absent.
const (
	FormatJSON Format = ".json"
	FormatYAML Format = ".yaml"
	FormatTOML Format = ".toml"
)

// undetermined is the zero Format: discovery does not narrow by extension.
const undetermined Format = ""

// errUnknownFormat is the refusal for a Format that is not one of the
// three declared constants.
var errUnknownFormat = errors.New("format must be spelled by one full extension")

// acceptedFormats renders the accepted format spellings for a refusal.
const acceptedFormats = `".json", ".yaml", ".toml"`

// check refuses a Format that no constant spells, naming the accepted
// spellings; discovery never guesses what an invented format means.
func (f Format) check() error {
	switch f {
	case FormatJSON, FormatYAML, FormatTOML, undetermined:
		return nil
	default:
		return fmt.Errorf("%w: got %q, accepted: %s", errUnknownFormat, string(f), acceptedFormats)
	}
}

// IsUndetermined reports whether the format narrows the search at all.
func (f Format) IsUndetermined() bool {
	return f == undetermined
}

// formatOf maps one file extension to its format; the empty extension
// yields the undetermined format.
func formatOf(ext Extension) Format {
	switch Format(ext) {
	case FormatJSON:
		return FormatJSON
	case FormatYAML:
		return FormatYAML
	case FormatTOML:
		return FormatTOML
	case undetermined:
		return undetermined
	default:
		return undetermined
	}
}
