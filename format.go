package goconf

import (
	"github.com/adambarber/gorefuse"
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

// acceptedFormats renders the accepted format spellings for a refusal.
const acceptedFormats = `".json", ".yaml", ".toml"`

// check refuses a Format that no constant spells, naming the accepted
// spellings; discovery never guesses what an invented format means.
func (f Format) check() error {
	switch f {
	case FormatJSON, FormatYAML, FormatTOML, undetermined:
		return nil
	default:
		return refuseWhy(ErrUnknownFormat,
			gorefuse.Subject("format \""+string(f)+"\""),
			gorefuse.Why("a format must be spelled by one full extension; accepted: "+acceptedFormats+"."))
	}
}

// IsUndetermined reports whether the format narrows the search at all.
func (f Format) IsUndetermined() bool {
	return f == undetermined
}

// formatOf maps one file extension to its format; the empty extension
// yields the undetermined format and false. An extension no constant
// spells yields false: the caller must refuse what it cannot spell.
func formatOf(ext Extension) (Format, bool) {
	switch Format(ext) {
	case FormatJSON:
		return FormatJSON, true
	case FormatYAML:
		return FormatYAML, true
	case FormatTOML:
		return FormatTOML, true
	case undetermined:
		return undetermined, true
	default:
		return undetermined, false
	}
}
