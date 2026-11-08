package goconf

import "strings"

// Unreachable through the public API; see docs/load.md. It rewrites
// one yaml terror line so the echoed scalar never reaches the refusal.
func stripYAMLTypeLine(msg refusalText) refusalText {
	head, tag, ok := strings.Cut(string(msg), terrorHead)
	if !ok {
		return stripBackquotedValues(msg)
	}

	return refusalText(head + terrorHead + string(cutTypeSpan(terrorTail(tag))))
}

// cutTypeSpan drops the echoed scalar from one terror tail: the tag
// token is kept, every backquoted span (balanced or trailing-open) goes,
// and the `into TYPE` clause survives untouched.
func cutTypeSpan(tag terrorTail) terrorTail {
	return terrorTail(stripBackquotedValues(refusalText(tag)))
}

// terrorTail is the part of a yaml terror line after its head.
type terrorTail string

// terrorHead begins every value-echoing yaml type error line.
const terrorHead = "cannot unmarshal "
