package goconf

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/adambarber/gorefuse"
)

// lineAtByte returns the 1-based line number of the byte at offset in data.
// JSON errors report byte offsets; refusals name lines, so the library
// translates at the boundary instead of letting offsets leak to callers.
func lineAtByte(data SourceBytes, offset int64) int {
	if offset < 0 {
		return 1
	}

	if offset > int64(len(data)) {
		offset = int64(len(data))
	}

	return bytes.Count(data[:offset], []byte("\n")) + 1
}

// jsonErr converts a decode error into a refusal naming the line the
// decoder stopped on, using the raw bytes it was reading. An unknown-field
// refusal names the key's line, found by re-walking the token stream.
func jsonErr(data SourceBytes, file FilePath, err error) error {
	var syntax *json.SyntaxError

	var typeErr *json.UnmarshalTypeError

	key, unknown := unknownFieldOf(err)

	switch {
	case errors.As(err, &syntax):
		line := lineAtByte(data, syntax.Offset)

		return refuseDecoder(
			gorefuse.At(gorefuse.File(file), gorefuse.Line(line),
				gorefuse.Subject("JSON syntax"),
				syntaxCheck("bracket or quote")), err)
	case errors.As(err, &typeErr):
		line := lineAtByte(data, typeErr.Offset)

		return typeMismatchRefusal(file, line, ObjectKey(typeErr.Field),
			jsonWantKind(typeErr.Type), jsonFoundTerm(goTypeName(typeErr.Value)))
	case unknown:
		line := jsonKeyLine(data, key)

		return refuseAt(ErrUnknownField, file, gorefuse.Line(line),
			gorefuse.Subject(`"`+string(key)+`"`),
			gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
	default:
		if len(data) == 0 {
			return refuseDecoder(
				gorefuse.At(gorefuse.File(file), 0,
					gorefuse.Subject("JSON syntax"),
					gorefuse.Why("the file is empty; add settings or delete it.")), err)
		}

		return refuseDecoder(
			gorefuse.At(gorefuse.File(file), 0,
				gorefuse.Subject("JSON syntax"),
				syntaxCheck(syntaxWord("bracket or quote"))), err)
	}
}

// syntaxWord is one character class a syntax fix names.
type syntaxWord string

// syntaxCheck is the one-action fix for a syntax fault: it names the
// character class to look for, never the decoder.
func syntaxCheck(missing syntaxWord) gorefuse.Why {
	return gorefuse.Why("check the last line for a missing " + string(missing) + ".")
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

// jsonUnknownPrefix is encoding/json's unknown-field refusal prefix, up
// to the quoted key.
const jsonUnknownPrefix = "json: unknown field "

// unknownFieldOf extracts the key an unknown-field refusal names.
func unknownFieldOf(err error) (ObjectKey, bool) {
	rest, ok := strings.CutPrefix(err.Error(), jsonUnknownPrefix)
	if !ok {
		return "", false
	}

	name, _, _ := strings.Cut(strings.Trim(rest, `"`), `"`)

	return ObjectKey(name), name != ""
}

// jsonKeyLine walks the token stream and returns the line of the first
// occurrence of the key in key position; 0 when the walk cannot place
// it. Inside an object a non-expected string is a key.
func jsonKeyLine(data SourceBytes, key ObjectKey) int {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var inSeq bool

	expectVal := false

	for {
		tok, err := dec.Token()
		if err != nil {
			return 0
		}

		line, hit := jsonKeyHit(tok, key, expectVal, inSeq, lineAtByte(data, dec.InputOffset()))
		if hit {
			return line
		}

		expectVal, inSeq = advanceKeyWalk(tok, expectVal, inSeq)
	}
}

// jsonKeyHit reports whether this token names the key, and its line.
func jsonKeyHit(tok json.Token, key ObjectKey, expectVal, inSeq bool, line int) (int, bool) {
	k, isKey := tok.(string)

	return line, isKey && !expectVal && !inSeq && ObjectKey(k) == key
}

// advanceKeyWalk folds one token into the key-line walk's two bits.
func advanceKeyWalk(tok json.Token, expectVal, inSeq bool) (bool, bool) {
	if d, isDelim := tok.(json.Delim); isDelim {
		switch d {
		case '[':
			return false, true
		case ']':
			return false, false
		default:
			return false, inSeq
		}
	}

	if expectVal {
		return false, inSeq
	}

	_, isKey := tok.(string)

	return isKey && !inSeq, inSeq
}
