package goconf

import (
	"strings"

	"github.com/adambarber/gorefuse"
)

// refusalText is rendered text inside a refusal message.
type refusalText string

// renderFiles renders the candidate file names for a refusal message.
func renderFiles(name configName, cands []Extension) refusalText {
	parts := make([]string, 0, len(cands))
	for _, c := range cands {
		parts = append(parts, string(name)+string(c))
	}

	return refusalText(strings.Join(parts, ", "))
}

// renderPaths renders the searched directories for a refusal message.
func renderPaths(paths []DirPath) refusalText {
	parts := make([]string, 0, len(paths))
	for _, p := range paths {
		parts = append(parts, string(p))
	}

	return refusalText(strings.Join(parts, ", "))
}

// lastExt is the last dot-segment of one candidate extension: the part
// a user would rename (the .bak in app.json.bak).
func lastExt(ext Extension) Extension {
	dot := strings.LastIndex(string(ext), ".")

	return Extension(string(ext)[dot:])
}

// unknownCandidateRefusal refuses one extension no format spells, naming
// the extension and the file, with one action.
func unknownCandidateRefusal(p FilePath, ext Extension) error {
	return refuseWhy(ErrUnknownCandidate,
		gorefuse.Subject(`"`+string(lastExt(ext))+`" is not a format the app can read`),
		gorefuse.Why("delete or rename "+string(p)+"."))
}

// nothingFoundRefusal builds the empty-discovery refusal: the Reason
// already says no config file was found, so the Subject names only what
// it uniquely adds — the looked-for names, the paths, and one action.
func nothingFoundRefusal(d Discovery) error {
	cands := []Extension{Extension(FormatJSON), Extension(FormatYAML), Extension(FormatTOML)}
	if !d.Format.IsUndetermined() {
		cands = []Extension{Extension(d.Format)}
	}

	return refuseWhy(ErrNothingFound,
		gorefuse.Subject("the app looked for "+string(renderFiles(d.Name, cands))+" in: "+string(renderPaths(d.Search))),
		gorefuse.Why("create your config at one of those locations."))
}
