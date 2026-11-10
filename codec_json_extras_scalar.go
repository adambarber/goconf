package goconf

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/adambarber/gorefuse"
)

type jsonText string

func setJSONScalar(dst reflect.Value, v jsonText, file FilePath, path FieldPath) error {
	k := dst.Kind()
	if k == reflect.String {
		dst.SetString(string(v))

		return nil
	}

	if k == reflect.Interface {
		dst.Set(reflect.ValueOf(string(v)))

		return nil
	}

	return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), kindUserTerm(kindString))
}

func setJSONBool(dst reflect.Value, v bool, file FilePath, path FieldPath) error {
	switch dst.Kind() {
	case reflect.Bool:
		dst.SetBool(v)

		return nil
	case reflect.Interface:
		dst.Set(reflect.ValueOf(v))

		return nil
	case reflect.Invalid:
		return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), kindUserTerm(kindBool))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan,
		reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice, reflect.String, reflect.Struct, reflect.UnsafePointer:
		return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), kindUserTerm(kindBool))
	}

	return nil
}

func setJSONNumber(dst reflect.Value, n json.Number, file FilePath, path FieldPath) error {
	switch dst.Kind() {
	case reflect.Interface:
		return setJSONAnyNum(dst, n, file, path)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return applyJSONNumeric(n, file, path, dst, kindUserTerm(kindFloat), parseJSONInt)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return applyJSONNumeric(n, file, path, dst, kindUserTerm(kindFloat), parseJSONUint)
	case reflect.Float32, reflect.Float64:
		return applyJSONNumeric(n, file, path, dst, kindUserTerm(kindString), parseJSONFloat)
	case reflect.Invalid, reflect.Bool, reflect.Complex64, reflect.Complex128:
		return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), kindUserTerm(kindFloat))
	case reflect.Array, reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice, reflect.String, reflect.Struct, reflect.UnsafePointer:
		return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), kindUserTerm(kindFloat))
	}

	return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), kindUserTerm(kindFloat))
}

func applyJSONNumeric(n json.Number, file FilePath, path FieldPath, dst reflect.Value, got userTerm, apply func(jsonText, reflect.Value) bool) error {
	if apply(jsonText(n.String()), dst) {
		return nil
	}

	return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), got)
}

func parseJSONInt(s jsonText, dst reflect.Value) bool {
	return parseJSONInteger(s, dst, true)
}

func parseJSONUint(s jsonText, dst reflect.Value) bool {
	return parseJSONInteger(s, dst, false)
}

func parseJSONInteger(s jsonText, dst reflect.Value, signed bool) bool {
	bits := dst.Type().Bits()
	if signed {
		iv, err := strconv.ParseInt(string(s), 10, bits)
		if err != nil {
			return false
		}

		dst.SetInt(iv)

		return true
	}

	uv, err := strconv.ParseUint(string(s), 10, bits)
	if err != nil {
		return false
	}

	dst.SetUint(uv)

	return true
}

func parseJSONFloat(s jsonText, dst reflect.Value) bool {
	fv, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		return false
	}

	dst.SetFloat(fv)

	return true
}

func setJSONAnyNum(dst reflect.Value, n json.Number, file FilePath, path FieldPath) error {
	ns := n.String()
	if i64, err := strconv.ParseInt(ns, 10, 64); err == nil {
		dst.Set(reflect.ValueOf(i64))

		return nil
	}

	f64, parseErr := strconv.ParseFloat(ns, 64)
	if parseErr != nil {
		return jsonExtrasTypeMismatch(file, path, kindOf(dst.Type()), kindUserTerm(kindString))
	}

	dst.Set(reflect.ValueOf(f64))

	return nil
}

func jsonExtrasTypeMismatch(file FilePath, path FieldPath, want valueKind, got userTerm) error {
	seg := lastPathSeg(path)

	return refuseAt(ErrTypeMismatch, file, 0,
		gorefuse.Subject("the value of \""+string(seg)+"\" is "+string(got)+", not "+string(kindUserTerm(want))),
		gorefuse.Why("change the value of \""+string(seg)+"\" to "+string(kindUserTerm(want))+"."))
}

func lastPathSeg(path FieldPath) FieldPath {
	s := string(path)
	if _, after, ok := strings.CutLast(s, "."); ok {
		return FieldPath(after)
	}

	if s == "" {
		return FieldPath("<root>")
	}

	return path
}
