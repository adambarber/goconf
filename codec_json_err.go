package goconf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
// stopped on, using the raw bytes it was reading.
func jsonErr(data SourceBytes, err error) error {
	var syntax *json.SyntaxError

	var typeErr *json.UnmarshalTypeError

	switch {
	case errors.As(err, &syntax):
		return fmt.Errorf("json: line %d: %w", lineAtByte(data, syntax.Offset), syntax)
	case errors.As(err, &typeErr):
		return fmt.Errorf("json: line %d: %w: got %s for %s", lineAtByte(data, typeErr.Offset), errTypeMismatch, typeErr.Value, typeErr.Type)
	default:
		return fmt.Errorf("json: %w", err)
	}
}
