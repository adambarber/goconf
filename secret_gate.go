package goconf

import (
	"reflect"
	"slices"
	"strings"
	"unicode"

	"github.com/adambarber/gorefuse"
)

// secretDecision is the declaration state inherited down a config path.
type secretDecision uint8

const (
	secretUnmarked secretDecision = iota
	secretMarked
	secretCleared
)

// declaredSecret reads the one schema tag that classifies a value.
func declaredSecret(f *reflect.StructField) secretDecision {
	switch f.Tag.Get("secret") {
	case "true":
		return secretMarked
	case "false":
		return secretCleared
	default:
		return secretUnmarked
	}
}

// looksSecret reports whether one key segment names a secret-like value.
// Words are reduced to a stem so plural and compound spellings cannot slip
// past an exact-word check.
func looksSecret(seg metaSeg) bool {
	indicators := []FieldPath{"password", "pass", "passwd", "pwd", "secret", "token", "credential", "apikey", "privatekey", "api", "private"}
	words := secretWordsOf(seg)

	for i, word := range words {
		stem := singular(word)
		if slices.Contains(indicators, stem) {
			return true
		}

		if (stem == "api" || stem == "private") && i+1 < len(words) && singular(words[i+1]) == "key" {
			return true
		}
	}

	return false
}

// singular folds a plural word back to its stem so one indicator covers
// "token" and "tokens" without a second hand-maintained list.
func singular(word FieldPath) FieldPath {
	s := string(word)
	if len(s) > 3 && strings.HasSuffix(s, "ies") {
		return FieldPath(strings.TrimSuffix(s, "ies") + "y")
	}

	if len(s) > 2 && strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") {
		return FieldPath(strings.TrimSuffix(s, "s"))
	}

	return word
}

// secretWordsOf splits a key into lower-cased words at separators and
// camel-case boundaries, so "apiKey", "api_key", and "API KEY" agree.
func secretWordsOf(seg metaSeg) []FieldPath {
	var b strings.Builder

	var previous rune

	for _, r := range string(seg) {
		if unicode.IsUpper(r) && unicode.IsLower(previous) {
			b.WriteByte(' ')
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteByte(' ')
		}

		previous = r
	}

	fields := strings.Fields(b.String())

	out := make([]FieldPath, 0, len(fields))
	for _, field := range fields {
		out = append(out, FieldPath(field))
	}

	return out
}

// checkSecretDeclarations refuses every present key whose name looks secret
// but whose schema field carries no secret tag. The tag is the declaration;
// the heuristic only forces one to exist.
func checkSecretDeclarations(keys KeySet, t reflect.Type, file FilePath) error {
	for _, path := range keys.pathsList() {
		if err := checkOneSecretPath(path, t, file); err != nil {
			return err
		}
	}

	return nil
}

func checkOneSecretPath(path FieldPath, t reflect.Type, file FilePath) error {
	rest := metaSeg(path)
	current := t
	decision := secretUnmarked

	for rest != "" {
		seg, tail, _ := splitPathSeg(rest)
		if seg.isIdx {
			rest = tail
			continue
		}

		decision = nextSecretDecision(current, seg.name, decision)
		current = secretStepType(current, seg.name)

		if decision == secretUnmarked && looksSecret(seg.name) {
			return refuseAt(ErrNeedDeclaration, file, 1,
				gorefuse.Subject("\""+string(path)+"\""),
				gorefuse.Why("add a secret tag to the field that holds it."))
		}

		rest = tail
	}

	return nil
}

// applySecretTag lets one explicit tag replace an inherited decision; an
// untagged node keeps the decision it was handed.
func applySecretTag(declared, current secretDecision) secretDecision {
	if declared != secretUnmarked {
		return declared
	}

	return current
}

// nextSecretDecision applies an explicit tag at this node; an untagged child
// inherits its parent's decision rather than reopening the heuristic.
func nextSecretDecision(t reflect.Type, name metaSeg, current secretDecision) secretDecision {
	f, ok := fieldMatching(t, keySegment(name))
	if !ok {
		return current
	}

	return applySecretTag(declaredSecret(f), current)
}

// secretStepType follows one schema segment into a field or map value.
func secretStepType(t reflect.Type, name metaSeg) reflect.Type {
	t = elemType(t)

	if t.Kind() == reflect.Struct {
		if f, ok := fieldMatching(t, keySegment(name)); ok {
			return f.Type
		}
	}

	if t.Kind() == reflect.Map {
		return t.Elem()
	}

	return t
}
