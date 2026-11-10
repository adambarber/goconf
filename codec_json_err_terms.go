package goconf

import (
	"reflect"
	"strings"

	"github.com/adambarber/gorefuse"
)

// syntaxCheck is the one-action fix for a syntax fault: it names the
// character class to look for, never the decoder.
func syntaxCheck() gorefuse.Why {
	return gorefuse.Why("check the last line for a missing bracket or quote.")
}

// goTypeName is one Go type spelling a decoder reports.
type goTypeName string

// userTerm is the user's word for one value shape.
type userTerm string

// jsonWantKind maps one reported Go type to the shared value kind the
// type-mismatch refusal names: the kind comes from reflect, so defined
// types (type port int) classify as their base kind.
func jsonWantKind(t reflect.Type) valueKind {
	return kindOf(t)
}

// jsonFoundTerm maps one reported found-value word ("string", "number",
// "array", "bool", "object") to the user's term; the word can carry the
// scalar itself ("number 1.5"), so only the first word is classified.
func jsonFoundTerm(v goTypeName) userTerm {
	word, _, _ := strings.Cut(string(v), " ")

	return kindUserTerm(jsonWordKind(kindWord(word)))
}

// kindWord is one decoder's word for a value kind under classification.
type kindWord string

// jsonWordKind maps one reported found-value word to the shared value
// kind; an unrecognized word falls to the table kind.
func jsonWordKind(word kindWord) valueKind {
	switch word {
	case "string":
		return kindString
	case "number":
		return kindFloat
	case "bool":
		return kindBool
	case "array":
		return kindList
	default:
		return kindTable
	}
}
