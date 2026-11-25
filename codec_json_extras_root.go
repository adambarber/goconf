package goconf

import (
	"encoding/json/jsontext"
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

func jsonExtrasRootShape(file FilePath, kind jsontext.Kind, dst reflect.Value) error {
	found, ok := jsonRootKindName(kind)
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

func jsonRootKindName(kind jsontext.Kind) (goTypeName, bool) {
	switch kind {
	case jsontext.KindBeginObject:
		return rootFoundObject, true
	case jsontext.KindBeginArray:
		return rootFoundArray, true
	case jsontext.KindString:
		return rootFoundString, true
	case jsontext.KindNumber:
		return rootFoundNumber, true
	case jsontext.KindTrue, jsontext.KindFalse:
		return rootFoundBool, true
	case jsontext.KindNull:
		return rootFoundNull, true
	case jsontext.KindInvalid, jsontext.KindEndObject, jsontext.KindEndArray:
		return "", false
	}

	return "", false
}
