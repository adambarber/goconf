package goconf

import (
	"encoding/json/jsontext"
	"strings"
)

func isJSONRootKind(kind jsontext.Kind) bool {
	switch kind {
	case '{', '[', '"', '0', 't', 'f', 'n':
		return true
	case jsontext.KindInvalid, '}':
		return false
	case ']':
		return false
	}

	return false
}

func jsonPointerLastKey(p jsontext.Pointer) ObjectKey {
	if p == "" {
		return ""
	}

	segs := strings.Split(string(p), "/")
	if len(segs) == 0 {
		return ""
	}

	key := segs[len(segs)-1]
	key = strings.ReplaceAll(key, "~1", "/")
	key = strings.ReplaceAll(key, "~0", "~")

	return ObjectKey(key)
}

func jsonErrQuotedName(msg refusalText) ObjectKey {
	s := string(msg)

	_, after, ok := strings.Cut(s, "\"")
	if !ok {
		return ""
	}

	name, _, ok := strings.Cut(after, "\"")
	if !ok {
		return ""
	}

	return ObjectKey(name)
}

func jsonKindFoundName(kind jsontext.Kind) goTypeName {
	switch kind {
	case '"':
		return "string"
	case '0':
		return "number"
	case 't', 'f':
		return "bool"
	case '[':
		return "array"
	case '{':
		return "object"
	case 'n':
		return "null"
	case jsontext.KindInvalid, '}', ']':
		return ""
	}

	return ""
}
