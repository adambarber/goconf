package goconf

import (
	"encoding/json/jsontext"
	"strings"
)

func isJSONRootKind(kind jsontext.Kind) bool {
	switch kind {
	case jsontext.KindBeginObject, jsontext.KindBeginArray, jsontext.KindString, jsontext.KindNumber,
		jsontext.KindTrue, jsontext.KindFalse, jsontext.KindNull:
		return true
	case jsontext.KindInvalid, jsontext.KindEndObject, jsontext.KindEndArray:
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
	case jsontext.KindString:
		return "string"
	case jsontext.KindNumber:
		return "number"
	case jsontext.KindTrue, jsontext.KindFalse:
		return "bool"
	case jsontext.KindBeginArray:
		return "array"
	case jsontext.KindBeginObject:
		return "object"
	case jsontext.KindNull:
		return "null"
	case jsontext.KindInvalid, jsontext.KindEndObject, jsontext.KindEndArray:
		return ""
	}

	return ""
}
