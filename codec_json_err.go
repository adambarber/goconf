package goconf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

// errTypeMismatch is the refusal for a value that does not fit its field.
var errTypeMismatch = errors.New("cannot unmarshal into Go field of declared type")

// jsonErr converts a decode error into one that names the line the decoder
// stopped on, using the raw bytes it was reading. An unknown-field refusal
// names the key's line, found by re-walking the token stream.
func jsonErr(data SourceBytes, err error) error {
	var syntax *json.SyntaxError

	var typeErr *json.UnmarshalTypeError

	key, unknown := unknownFieldOf(err)

	switch {
	case errors.As(err, &syntax):
		return fmt.Errorf("json: line %d: %w", lineAtByte(data, syntax.Offset), syntax)
	case errors.As(err, &typeErr):
		return fmt.Errorf("json: line %d: %w: got %s for %s", lineAtByte(data, typeErr.Offset), errTypeMismatch, typeErr.Value, typeErr.Type)
	case unknown:
		return fmt.Errorf("json: line %d: %w %q", jsonKeyLine(data, key), errUnknownFieldJSON, string(key))
	default:
		return fmt.Errorf("json: %w", err)
	}
}

// errUnknownFieldJSON is the refusal for a key the target does not
// declare, with the key's line named by the caller's wrap.
var errUnknownFieldJSON = errors.New("unknown field")

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
