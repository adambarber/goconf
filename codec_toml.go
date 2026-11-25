package goconf

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adambarber/gorefuse"
)

// decodeTOMLStrict decodes TOML bytes into the target, refusing unknown
// fields and type mismatches with the offending field named.
func decodeTOMLStrict(data SourceBytes, file FilePath, dst Target) error {
	shadowType := tomlShadowType(dst.Value.Type())
	if shadowType == dst.Value.Type() {
		md, err := toml.NewDecoder(bytes.NewReader(data)).Decode(dst.Value.Addr().Interface())
		if err != nil {
			return tomlDecodeRefusal(file, err)
		}

		return tomlUndecoded(dst, data, &md, file, nil)
	}

	shadow := reflect.New(shadowType).Elem()

	md, err := toml.NewDecoder(bytes.NewReader(data)).Decode(shadow.Addr().Interface())
	if err != nil {
		return tomlDecodeRefusal(file, err)
	}

	absorbed := fillValueShadows(dst.Value, shadow, &md, data, file, toml.Key{})

	return tomlUndecoded(dst, data, &md, file, absorbed)
}

// tomlDecodeRefusal classifies one BurntSushi error at the boundary: a
// ParseError is a syntax fault, an "incompatible types" message is
// ErrTypeMismatch; decode errors are untyped, so shapes are matched.
func tomlDecodeRefusal(file FilePath, err error) error {
	msg := tomlMessage(err.Error())

	if msg.has("incompatible types:") || msg.has("type mismatch for ") {
		return typeMismatchRefusal(file, msg.line(), ObjectKey(msg.key().unquote()),
			tomlWantKind(msg), tomlFoundTerm(msg))
	}

	return tomlRefusal(file, err)
}

// has reports whether the message contains the span.
func (m tomlMessage) has(s spanText) bool { return strings.Contains(string(m), string(s)) }

// key is the key the message names; the shape is
// 'toml: line N (last key "K"): ...'.
func (m tomlMessage) key() keyQuoted {
	_, rest, ok := strings.Cut(string(m), `(last key "`)
	if !ok {
		return "the offending value"
	}

	key, _, _ := strings.Cut(rest, `")`)

	return keyQuoted(`"` + key + `"`)
}

// line is the line the message names; 0 when none.
func (m tomlMessage) line() int {
	var line int

	if _, err := fmt.Sscanf(string(m), "toml: line %d", &line); err == nil && line > 0 {
		return line
	}

	return 0
}

// tomlWantKind maps one message's "destination has type T" clause to the
// shared value kind; the "expected table but found" shape already names
// a table, and any unrecognized destination falls to the table kind.
func tomlWantKind(m tomlMessage) valueKind {
	_, rest, ok := strings.Cut(string(m), "destination has type ")
	if !ok {
		return kindTable
	}

	word, _, _ := strings.Cut(rest, " ")
	word = strings.TrimSuffix(word, ".")

	return tomlWordKind(kindWord(word))
}

// tomlWordKind maps one "destination has type" word to the shared value
// kind; an unrecognized word falls to the table kind. BurntSushi spells the
// list destination "slice", sometimes "[]T" with the bracket included.
func tomlWordKind(word kindWord) valueKind {
	if word == "slice" || strings.HasPrefix(string(word), "[]") {
		return kindList
	}

	switch word {
	case "integer":
		return kindInt
	case "float":
		return kindFloat
	case tomlWordString:
		return kindString
	case "boolean":
		return kindBool
	default:
		return kindTable
	}
}

// tomlFoundTerm maps one message's "TOML value has type T" clause to the
// user's term; empty when the message names none.
func tomlFoundTerm(m tomlMessage) userTerm {
	_, rest, ok := strings.Cut(string(m), "TOML value has type ")
	if !ok {
		return ""
	}

	word, _, _ := strings.Cut(rest, ";")
	word = strings.TrimSpace(word)

	switch word {
	case tomlWordString:
		return kindUserTerm(kindString)
	case "integer", "int64":
		return kindUserTerm(kindInt)
	case "float", "float64":
		return kindUserTerm(kindFloat)
	case "boolean":
		return kindUserTerm(kindBool)
	default:
		return ""
	}
}

// tomlWordString is BurntSushi's spelling of the string type in both
// type-clause shapes.
const tomlWordString = "string"

// tomlRefusal rewrites one BurntSushi parse refusal as the
// ErrDecoderRefused carrier: the stripped detail survives as the cause.
func tomlRefusal(file FilePath, err error) error {
	return refuseDecoder(
		gorefuse.At(gorefuse.File(file), gorefuse.Line(tomlParseLine(err)),
			gorefuse.Subject("TOML syntax"),
			gorefuse.Why("check the last line for a missing bracket or quote.")), errTOMLValueFree(err))
}

// tomlParseLine is the line one BurntSushi ParseError names; 0 when none.
func tomlParseLine(err error) int {
	if perr, ok := errors.AsType[toml.ParseError](err); ok {
		return perr.Position.Line
	}

	return 0
}
