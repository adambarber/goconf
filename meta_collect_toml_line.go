package goconf

import (
	"strings"

	"github.com/BurntSushi/toml"
)

type tomlLineText metaSeg

func tomlFindKeyLine(src SourceBytes, cursor int, k toml.Key, md *toml.MetaData) (int, int) {
	if len(k) == 0 {
		return 1, cursor
	}

	if idx, line := tomlFindKeyLineHeuristic(src, cursor, k, md); idx >= 0 {
		return line, idx
	}

	if idx, line := tomlFindKeyLineHeuristic(src, 0, k, md); idx >= 0 {
		return line, idx
	}

	return 1, cursor
}

func tomlFindKeyLineHeuristic(src SourceBytes, cursor int, k toml.Key, md *toml.MetaData) (int, int) {
	text := metaSeg(src)
	if cursor < 0 || cursor > len(text) {
		cursor = 0
	}

	full := metaSeg(strings.Join(k, "."))
	leaf := metaSeg(k[len(k)-1])
	typ := metaSeg(md.Type(k...))
	lines := strings.Split(string(text[cursor:]), "\n")
	offset := cursor

	for i, ln := range lines {
		lineNo := 1 + strings.Count(string(text[:offset]), "\n")
		trimmed := trimTOMLLine(tomlLineText(ln))

		if tomlLineMatchesKey(trimmed, full, leaf, typ) {
			return offset, lineNo
		}

		offset += len(ln)
		if i < len(lines)-1 {
			offset++
		}
	}

	return -1, 0
}

func trimTOMLLine(ln tomlLineText) tomlLineText {
	out := []rune{}
	inSingle := false
	inDouble := false

	for _, r := range string(ln) {
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle

			out = append(out, r)
		case r == '"' && !inSingle:
			inDouble = !inDouble

			out = append(out, r)
		case r == '#' && !inSingle && !inDouble:
			return tomlLineText(strings.TrimSpace(string(out)))
		default:
			out = append(out, r)
		}
	}

	return tomlLineText(strings.TrimSpace(string(out)))
}

func tomlLineMatchesKey(trimmed tomlLineText, full, leaf, typ metaSeg) bool {
	if trimmed == "" {
		return false
	}

	if typ == metaSeg(tomlTypeArrayHash) {
		return tomlMatchTableHeader(trimmed, full, leaf, "[[", "]]")
	}

	if typ == metaSeg("Hash") {
		return tomlMatchTableHeader(trimmed, full, leaf, "[", "]")
	}

	return tomlMatchAssignment(trimmed, full) || tomlMatchAssignment(trimmed, leaf)
}

func tomlMatchTableHeader(trimmed tomlLineText, full, leaf, open, closer metaSeg) bool {
	if tomlHasHeader(trimmed, full, open, closer) {
		return true
	}

	return tomlHasHeader(trimmed, leaf, open, closer)
}

func tomlHasHeader(trimmed tomlLineText, key, open, closer metaSeg) bool {
	h := open + key + closer

	return tomlLineText(h) == trimmed || strings.HasPrefix(string(trimmed), string(h))
}

func tomlMatchAssignment(trimmed tomlLineText, key metaSeg) bool {
	if tomlIsAssignmentLine(trimmed, key) {
		return true
	}

	return tomlIsQuotedAssignmentLine(trimmed, key)
}

func tomlIsAssignmentLine(trimmed tomlLineText, key metaSeg) bool {
	if !strings.HasPrefix(string(trimmed), string(key)) {
		return false
	}

	rest := strings.TrimSpace(string(trimmed)[len(key):])

	return strings.HasPrefix(rest, "=")
}

func tomlIsQuotedAssignmentLine(trimmed tomlLineText, key metaSeg) bool {
	for _, q := range []metaSeg{`"`, `'`} {
		qk := q + key + q
		if strings.HasPrefix(string(trimmed), string(qk)) {
			rest := strings.TrimSpace(string(trimmed)[len(qk):])

			return strings.HasPrefix(rest, "=")
		}
	}

	return false
}
