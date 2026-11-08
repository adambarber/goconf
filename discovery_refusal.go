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

// nothingFoundRefusal builds the empty-discovery refusal: it names the
// file name(s) the app looked for and the directories it looked in, and
// gives one action — create the config at one of those locations.
func nothingFoundRefusal(d Discovery) error {
	cands := []Extension{Extension(FormatJSON), Extension(FormatYAML), Extension(FormatTOML)}
	if !d.Format.IsUndetermined() {
		cands = []Extension{Extension(d.Format)}
	}

	return refuseWhy(ErrNothingFound,
		gorefuse.Subject("no config file was found"),
		gorefuse.Why("the app looked for "+string(renderFiles(d.Name, cands))+" in: "+string(renderPaths(d.Search))+
			". create your config at one of those locations."))
}
