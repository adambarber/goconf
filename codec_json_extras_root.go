package goconf

import (
	"encoding/json"
	"reflect"
)

const (
	rootFoundObject goTypeName = "object"
	rootFoundArray  goTypeName = "array"
	rootFoundString goTypeName = "string"
	rootFoundNumber goTypeName = "number"
	rootFoundBool   goTypeName = "bool"
	rootFoundNull   goTypeName = "null"
)

func jsonExtrasRootShape(file FilePath, tok json.Token, dst reflect.Value) error {
	found, ok := jsonRootTokenKind(tok)
	if !ok {
		return nil
	}

	switch dst.Kind() {
	case reflect.Struct, reflect.Map:
		if found != rootFoundObject {
			return jsonRootShapeRefusal(file, 0, found)
		}
	case reflect.Slice, reflect.Array:
		if found != rootFoundArray {
			return jsonRootShapeRefusal(file, 0, found)
		}
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Pointer, reflect.String, reflect.UnsafePointer:
		return nil
	}

	return nil
}

func jsonRootTokenKind(tok json.Token) (goTypeName, bool) {
	switch d := tok.(type) {
	case json.Delim:
		if d == '{' {
			return rootFoundObject, true
		}

		if d == '[' {
			return rootFoundArray, true
		}

		return "", false
	case string:
		return rootFoundString, true
	case json.Number:
		return rootFoundNumber, true
	case bool:
		return rootFoundBool, true
	case nil:
		return rootFoundNull, true
	default:
		return "", false
	}
}
