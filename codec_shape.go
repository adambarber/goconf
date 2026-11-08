package goconf

import (
	"github.com/adambarber/gorefuse"
)

// valueKind is one user-shaped value kind a decoder reports: the single
// vocabulary every format's type-mismatch refusal synthesizes through.
type valueKind string

const (
	kindInt    valueKind = "int"
	kindFloat  valueKind = "float"
	kindString valueKind = "string"
	kindBool   valueKind = "bool"
	kindList   valueKind = "list"
	kindTable  valueKind = "table"
)

// kindUserTerm is the one helper mapping a value kind to the user's word
// for that shape: no per-format copies of the vocabulary.
func kindUserTerm(k valueKind) userTerm {
	switch k {
	case kindInt:
		return "a whole number"
	case kindFloat:
		return "a decimal number"
	case kindString:
		return "text"
	case kindBool:
		return "yes/no"
	case kindList:
		return "a list"
	case kindTable:
		return "a table"
	default:
		return "a value of the app's own shape"
	}
}

// typeMismatchRefusal builds the type-mismatch refusal naming the key
// and the user's word for the expected shape; the Reason renders
// "type mismatch", so the subject must not repeat it.
func typeMismatchRefusal(file FilePath, line int, key ObjectKey, want valueKind, found userTerm) error {
	what := "the value"

	if key != "" {
		what = `the value of "` + string(key) + `"`
	}

	subject := gorefuse.Subject(what + " is not " + string(kindUserTerm(want)))
	if found != "" {
		subject = gorefuse.Subject(what + " is " + string(found) + ", not " + string(kindUserTerm(want)))
	}

	return refuseAt(ErrTypeMismatch, file, gorefuse.Line(line), subject,
		gorefuse.Why("change "+what+" to "+string(kindUserTerm(want))+"."))
}
