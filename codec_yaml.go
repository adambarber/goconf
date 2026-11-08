package goconf

import (
	"errors"
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
			return refuseReason(r,
				gorefuse.At(gorefuse.File(file), 0, refusalSubject(r), refusalWhyText(r)),
				errValueFree(err))
		}
	}

	return decodeRefusal(file, errValueFree(err))
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

// refusalSubject names the fault for the classified YAML refusals.
func refusalSubject(r Reason) gorefuse.Subject {
	if r == ErrUnknownField {
		return gorefuse.Subject("YAML unknown field")
	}

	return gorefuse.Subject("YAML type mismatch")
}

// refusalWhyText is the remediation for the classified YAML refusals.
func refusalWhyText(r Reason) gorefuse.Why {
	if r == ErrUnknownField {
		return gorefuse.Why("the decode target does not declare them; declare the fields in the target or remove them from the file.")
	}

	return gorefuse.Why("the value does not fit the field's declared type; fix the value.")
}

// decodeRefusal rewrites one decoder error as the ErrDecoderRefused
// carrier: the stripped detail survives as the cause tail.
func decodeRefusal(file FilePath, err error) error {
	return refuseDecoder(
		gorefuse.At(gorefuse.File(file), 0,
			gorefuse.Subject("YAML decode"),
			gorefuse.Why("fix the error the YAML decoder reported.")), err)
}

// multiDocRefusal refuses a stream holding more than one document, naming
// the separator's line.
func multiDocRefusal(file FilePath, line int) error {
	return refuseAt(ErrMultiDoc, file, gorefuse.Line(line),
		gorefuse.Subject("second document separator"),
		gorefuse.Why("first-wins would silently resolve the stream; hold one document per file."))
}
