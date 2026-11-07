package goconf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// frame is one open container on the token-stack walk: its dotted path and
// whether its members are values of a sequence rather than object keys.
type frame struct {
	path  FieldPath
	isSeq bool
}

// walk is the mutable state of the token-stream key walk.
type walk struct {
	pending   FieldPath
	stack     []frame
	expectVal bool
}

// errUnbalanced is the refusal for a close delimiter past the root.
var errUnbalanced = errors.New("unbalanced delimiter")

// errTrailing is the refusal for bytes after the document root.
var errTrailing = errors.New("trailing data after the document")

// under joins a key to its parent's dotted path, records it, and returns
// the composed path, so an opening delimiter can open a frame at the child
// path.
func under(keys *KeySet, parent, key FieldPath) FieldPath {
	p := key
	if parent != "" {
		p = parent + "." + key
	}

	*keys = keys.Add(p)

	return p
}

// decodeJSONStrict decodes JSON bytes into the target, refusing unknown
// fields, trailing data, and type mismatches with the line named.
func decodeJSONStrict(data SourceBytes, dst Target) error {
	if err := refuseDuplicateJSONKeys(data); err != nil {
		return err
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst.Value.Addr().Interface()); err != nil {
		return jsonErr(data, err)
	}

	if _, err := dec.Token(); err == nil {
		return fmt.Errorf("json: %w", errTrailing)
	}

	return nil
}

// the JSON document, walking the token stream so presence is recorded no
// matter how deeply it is nested, including inside arrays of objects.
func presentJSONKeys(data SourceBytes) (KeySet, error) {
	keys := KeySet{paths: nil}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	w := walk{stack: []frame{{path: "", isSeq: false}}, pending: "", expectVal: false}

	for {
		tok, err := dec.Token()
		switch {
		case err == io.EOF:
			return keys, nil
		case err != nil:
			return keys, jsonErr(data, err)
		default:
			line := lineAtByte(data, dec.InputOffset())
			if inner := advanceToken(&keys, &w, tok, line); inner != nil {
				return keys, inner
			}
		}
	}
}

// advanceToken folds one token into the key walk. expectVal is true while
// the next token is the value of an already-recorded key; any token then
// consumes it. Outside it, a string token in an object is a key.
func advanceToken(keys *KeySet, w *walk, tok json.Token, line int) error {
	top := w.stack[len(w.stack)-1]
	if delim, isDelim := tok.(json.Delim); isDelim {
		return advanceDelim(w, delim, line)
	}

	if w.expectVal {
		w.expectVal = false
		return nil
	}

	if key, isKey := tok.(string); isKey && !top.isSeq {
		w.pending = under(keys, top.path, FieldPath(key))
		w.expectVal = true
	}

	return nil
}

// advanceDelim handles one structural token: open pushes a frame at the
// pending key's path, close pops one. A close that would empty the walk
// past the document root is refused.
func advanceDelim(w *walk, delim json.Delim, line int) error {
	open := frame{path: w.pending, isSeq: false}

	switch delim {
	case '{':
	case '[':
		open.isSeq = true
	default:
		if len(w.stack) <= 1 {
			return fmt.Errorf("json: line %d: %w", line, errUnbalanced)
		}

		return closeFrame(w)
	}

	w.stack = append(w.stack, open)

	return closeExpect(w)
}

// closeFrame pops the current frame and clears its pending state.
func closeFrame(w *walk) error {
	w.stack = w.stack[:len(w.stack)-1]

	return closeExpect(w)
}

// closeExpect clears the value-expectation and pending path between tokens.
func closeExpect(w *walk) error {
	w.expectVal = false

	w.pending = ""

	return nil
}
