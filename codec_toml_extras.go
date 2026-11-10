package goconf

import (
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/adambarber/gorefuse"
)

func tomlRouteExtras(dst Target, data SourceBytes, undec []toml.Key, file FilePath) error {
	raw, err := decodeTOMLRaw(data, file)
	if err != nil {
		return err
	}

	handled := map[string]bool{}

	var unknown []toml.Key

	for _, k := range undec {
		ks := k.String()
		if handled[ks] {
			continue
		}

		ok, err := tomlCaptureKey(dst.Value, raw, k, file)
		if err != nil {
			return err
		}

		if ok {
			handled[ks] = true
		} else {
			unknown = append(unknown, k)
		}
	}

	if len(unknown) > 0 {
		return refuseAt(ErrUnknownField, file, 0,
			gorefuse.Subject(string(joinKeys(unknown))),
			gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
	}

	return nil
}

func tomlCaptureKey(dst, raw reflect.Value, k toml.Key, file FilePath) (bool, error) {
	if len(k) == 0 {
		return false, nil
	}

	v := derefValue(dst)
	if !v.IsValid() {
		return false, nil
	}

	return tomlCaptureAt(v, raw, k, 0, file)
}

func tomlCaptureAt(v, raw reflect.Value, k toml.Key, idx int, file FilePath) (bool, error) {
	if idx >= len(k) {
		return false, nil
	}

	seg := keySegment(k[idx])
	v = derefValue(v)

	if v.Kind() == reflect.Struct {
		return tomlCaptureStructAt(v, raw, k, idx, seg, file)
	}

	if v.Kind() == reflect.Map {
		return tomlCaptureMapAt(v, raw, k, idx, seg, file)
	}

	return false, nil
}

func tomlCaptureStructAt(v, raw reflect.Value, k toml.Key, idx int, seg keySegment, file FilePath) (bool, error) {
	sf, ok := findStructFieldByKey(v.Type(), seg)
	if ok {
		return tomlCaptureAt(v.FieldByIndex(sf.Index), rawChildRV(raw, seg), k, idx+1, file)
	}

	ef, ok := extrasField(v.Type())
	if !ok {
		return false, nil
	}

	return tomlSetExtra(v.FieldByIndex(ef.Index), raw, k, idx, file)
}

func tomlCaptureMapAt(v, raw reflect.Value, k toml.Key, idx int, seg keySegment, file FilePath) (bool, error) {
	mv := mapGetStringKey(v, seg)
	if !mv.IsValid() {
		return false, nil
	}

	return tomlCaptureAt(mv, rawChildRV(raw, seg), k, idx+1, file)
}

func tomlSetExtra(extras, raw reflect.Value, k toml.Key, idx int, file FilePath) (bool, error) {
	if extras.Kind() != reflect.Map {
		return false, nil
	}

	if extras.IsNil() && extras.CanSet() {
		extras.Set(reflect.MakeMap(extras.Type()))
	}

	key := reflect.ValueOf(k[idx]).Convert(extras.Type().Key())

	val := rawChildRV(raw, keySegment(k[idx]))
	if err := setExtraValue(extras, key, val, file, FieldPath(k[idx])); err != nil {
		return false, err
	}

	return true, nil
}

func setExtraValue(extras, key, val reflect.Value, file FilePath, path FieldPath) error {
	elemType := extras.Type().Elem()

	if !val.IsValid() {
		extras.SetMapIndex(key, reflect.Zero(elemType))
		return nil
	}

	if val.Type().ConvertibleTo(elemType) {
		extras.SetMapIndex(key, val.Convert(elemType))
		return nil
	}

	if elemType.Kind() == reflect.Interface && elemType.NumMethod() == 0 {
		extras.SetMapIndex(key, val)
		return nil
	}

	if val.Type().AssignableTo(elemType) {
		extras.SetMapIndex(key, val)
		return nil
	}

	return typeMismatchRefusal(file, 0, ObjectKey(path), kindOf(elemType), tomlRawValueTerm(val))
}

func tomlRawValueTerm(v reflect.Value) userTerm {
	if !v.IsValid() {
		return ""
	}

	for v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map, reflect.Struct:
		return kindUserTerm(kindTable)
	case reflect.Slice, reflect.Array:
		return kindUserTerm(kindList)
	case reflect.Bool:
		return kindUserTerm(kindBool)
	case reflect.String:
		return kindUserTerm(kindString)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return kindUserTerm(kindFloat)
	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Pointer, reflect.UnsafePointer:
		return ""
	}

	return ""
}

func decodeTOMLRaw(data SourceBytes, file FilePath) (reflect.Value, error) {
	rv := reflect.New(reflect.TypeOf((*interface{})(nil)).Elem())
	if _, err := toml.Decode(string(data), rv.Interface()); err != nil {
		return reflect.Value{}, tomlDecodeRefusal(file, err)
	}

	return rv.Elem(), nil
}
