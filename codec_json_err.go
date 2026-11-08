package goconf

import (
	"bytes"
	"encoding/json"
	"errors"
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
				gorefuse.Why("fix the JSON syntax in your config file; the line is where reading stopped.")), err)
	case errors.As(err, &typeErr):
		line := lineAtByte(data, typeErr.Offset)

		return refuseAt(ErrTypeMismatch, file, gorefuse.Line(line),
			valueShapeSubject(goTypeName(typeErr.Value), goTypeName(typeErr.Type.String())),
			valueShapeWhy(goTypeName(typeErr.Type.String())))
	case unknown:
		line := jsonKeyLine(data, key)

		return refuseAt(ErrUnknownField, file, gorefuse.Line(line),
			gorefuse.Subject(`"`+string(key)+`"`),
			gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
	default:
		return refuseDecoder(
			gorefuse.At(gorefuse.File(file), 0,
				gorefuse.Subject("JSON syntax"),
				gorefuse.Why("fix the JSON syntax in your config file.")), err)
	}
}

// goTypeName is one Go type spelling a decoder reports.
type goTypeName string

// userTerm is the user's word for one value shape.
type userTerm string

// valueShapeSubject renders the user's value shape against what the app
// expects: "expected a number, found text", never the Go type. The
// Reason renders "type mismatch"; the subject must not repeat it.
func valueShapeSubject(found, want goTypeName) gorefuse.Subject {
	return gorefuse.Subject("expected " + string(goTypeToUserTerm(want)) + ", found " + string(goTypeToUserTerm(found)))
}

// valueShapeWhy is the one-action fix for a type mismatch.
func valueShapeWhy(want goTypeName) gorefuse.Why {
	return gorefuse.Why("change the value to " + string(goTypeToUserTerm(want)) + ".")
}

// goTypeToUserTerm maps one Go type spelling to the user's word for it.
// Every returned term is self-articulated ("a number", "text").
func goTypeToUserTerm(t goTypeName) userTerm {
	switch t {
	case "string", "text":
		return "text"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "number":
		return "a number"
	case "bool", "boolean":
		return "true or false"
	case "[]interface {}", "array":
		return "a list"
	case "map[string]interface {}", "object":
		return "a table"
	default:
		if strings.Contains(string(t), ".") {
			return "a value of the app's own shape"
		}

		return "a " + userTerm(t)
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
