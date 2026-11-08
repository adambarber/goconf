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
				gorefuse.Why("fix the syntax the decoder stopped on; the frame names its line.")), err)
	case errors.As(err, &typeErr):
		line := lineAtByte(data, typeErr.Offset)

		return refuseAt(ErrTypeMismatch, file, gorefuse.Line(line),
			gorefuse.Subject("value of type \""+typeErr.Value+"\" for a field of type "+typeErr.Type.String()),
			gorefuse.Why("the value does not fit the field's declared type; the frame names the offending line."))
	case unknown:
		line := jsonKeyLine(data, key)

		return refuseAt(ErrUnknownField, file, gorefuse.Line(line),
			gorefuse.Subject("\""+string(key)+"\""),
			gorefuse.Why("the decode target does not declare it; declare the field in the target or remove it from the file."))
	default:
		return refuseDecoder(
			gorefuse.At(gorefuse.File(file), 0,
				gorefuse.Subject("JSON decode"),
				gorefuse.Why("fix the error the JSON decoder reported.")), err)
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
