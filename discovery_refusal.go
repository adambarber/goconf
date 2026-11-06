package goconf

import "strings"

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
