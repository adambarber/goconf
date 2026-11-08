package goconf

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/adambarber/gorefuse"
	"gopkg.in/yaml.v3"
)

// decodeYAMLStrict decodes YAML bytes into the target, refusing unknown
// fields and type mismatches. Explicit null is accepted and decodes to the
// field's zero value: it is a decision, not an absence.
func decodeYAMLStrict(data SourceBytes, file FilePath, dst Target) error {
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)

	if err := dec.Decode(dst.Value.Addr().Interface()); err != nil {
		return yamlDecodeRefusal(file, err)
	}

	var extra yaml.Node
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return multiDocRefusal(file, extra.Line)
	}

	return nil
}

// yamlUnknownFieldTail is the shape of yaml.v3's unknown-field terror:
// KnownFields(true) emits "line N: field KEY not found in type T". The
// error has no typed discriminator, so classification matches this shape.
const yamlUnknownFieldTail = " not found in type "

// yamlDecodeRefusal classifies one yaml.v3 decode error: all-unknown-field
// TypeError → ErrUnknownField, other TypeError → ErrTypeMismatch,
// scanner/parser faults → ErrDecoderRefused; detail survives as the cause.
func yamlDecodeRefusal(file FilePath, err error) error {
	typeErr, ok := errors.AsType[*yaml.TypeError](err)
	if ok {
		r := yamlTypeReason(typeErr)

		if r != ErrDecoderRefused {
			if r == ErrUnknownField {
				key, _ := strings.CutPrefix(string(yamlUnknownFieldOf(typeErr)), "field ")

				return refuseAt(ErrUnknownField, file, gorefuse.Line(yamlTypeLine(typeErr)),
					gorefuse.Subject(`"`+key+`"`),
					gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
			}

			return refuseAt(ErrTypeMismatch, file, gorefuse.Line(yamlTypeLine(typeErr)),
				gorefuse.Subject("expected a number, found text"),
				gorefuse.Why("change the value to a number."))
		}
	}

	return decodeRefusal(file, errValueFree(err))
}

// yamlTypeLine is the first line any of the TypeError's issues names; 0
// when none carries a line.
func yamlTypeLine(typeErr *yaml.TypeError) int {
	for _, e := range typeErr.Errors {
		var line int

		if _, err := fmt.Sscanf(e, "line %d:", &line); err == nil && line > 0 {
			return line
		}
	}

	return 0
}

// yamlTypeReason classifies one TypeError by its terror lines: every
// issue naming "field K not found in type T" is an unknown field; other
// issues are type mismatches. A mix falls to ErrTypeMismatch.
func yamlTypeReason(typeErr *yaml.TypeError) Reason {
	if len(typeErr.Errors) == 0 {
		return ErrDecoderRefused
	}

	for _, e := range typeErr.Errors {
		if !strings.Contains(e, yamlUnknownFieldTail) {
			return ErrTypeMismatch
		}
	}

	return ErrUnknownField
}

// yamlFieldSpan is one field name an unknown-field terror names.
type yamlFieldSpan string

// yamlUnknownFieldOf is the key name one unknown-field terror names:
// the documented shape is "line N: field K not found in type T".
func yamlUnknownFieldOf(typeErr *yaml.TypeError) yamlFieldSpan {
	for _, e := range typeErr.Errors {
		if strings.Contains(e, yamlUnknownFieldTail) {
			_, rest, ok := strings.Cut(e, "line ")
			if !ok {
				continue
			}

			_, body, ok := strings.Cut(rest, ": ")
			if !ok {
				continue
			}

			key, _, _ := strings.Cut(body, yamlUnknownFieldTail)

			return yamlFieldSpan(strings.TrimPrefix(key, "field "))
		}
	}

	return ""
}

// decodeRefusal rewrites one scanner or parser fault as the
// ErrDecoderRefused carrier: the stripped detail survives as the cause
// tail.
func decodeRefusal(file FilePath, err error) error {
	return refuseDecoder(
		gorefuse.At(gorefuse.File(file), 0,
			gorefuse.Subject("YAML syntax"),
			gorefuse.Why("fix the YAML syntax in your config file.")), err)
}

// multiDocRefusal refuses a stream holding more than one document, naming
// the separator's line.
func multiDocRefusal(file FilePath, line int) error {
	return refuseAt(ErrMultiDoc, file, gorefuse.Line(line),
		gorefuse.Subject("second document separator"),
		gorefuse.Why("hold one document per file; only the first would be read."))
}
