package goconf

import (
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/adambarber/gorefuse"
)

func tomlRouteExtras(dst Target, data SourceBytes, undec []toml.Key, md *toml.MetaData, file FilePath) error {
	raw, err := decodeTOMLRaw(data, file)
	if err != nil {
		return err
	}

	aotIdx := tomlUndecodedAOTIndices(undec, md)

	var unknown []toml.Key

	for i, k := range undec {
		ok, err := tomlCaptureKey(dst.Value, raw, k, aotIdx[i], file)
		if err != nil {
			return err
		}

		if !ok {
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

func tomlCaptureKey(dst, raw reflect.Value, k toml.Key, aot tomlAOTIndex, file FilePath) (bool, error) {
	if len(k) == 0 {
		return false, nil
	}

	v := derefValue(dst)
	if !v.IsValid() {
		return false, nil
	}

	return tomlCaptureAt(v, raw, k, aot, 0, file)
}

func tomlCaptureAt(v, raw reflect.Value, k toml.Key, aot tomlAOTIndex, idx int, file FilePath) (bool, error) {
	if idx >= len(k) {
		return false, nil
	}

	seg := keySegment(k[idx])
	v = derefValue(v)

	if v.Kind() == reflect.Struct {
		return tomlCaptureStructAt(v, raw, k, aot, idx, seg, file)
	}

	if v.Kind() == reflect.Map {
		return tomlCaptureMapAt(v, raw, k, aot, idx, seg, file)
	}

	return false, nil
}

func tomlCaptureStructAt(v, raw reflect.Value, k toml.Key, aot tomlAOTIndex, idx int, seg keySegment, file FilePath) (bool, error) {
	sf, ok := findStructFieldByKey(v.Type(), seg)
	if ok {
		fieldVal := v.FieldByIndex(sf.Index)
		childRaw := rawChildRV(raw, seg)

		if indexed, indexedOK, err := tomlCaptureAOTIndexed(fieldVal, childRaw, k, aot, idx+1, file); indexed {
			return indexedOK, err
		}

		return tomlCaptureAt(fieldVal, childRaw, k, aot, idx+1, file)
	}

	ef, ok := extrasField(v.Type())
	if !ok {
		return false, nil
	}

	return tomlSetExtra(v.FieldByIndex(ef.Index), raw, k, idx, file)
}

func tomlCaptureMapAt(v, raw reflect.Value, k toml.Key, aot tomlAOTIndex, idx int, seg keySegment, file FilePath) (bool, error) {
	mapKey := reflect.ValueOf(string(seg)).Convert(v.Type().Key())

	mv := v.MapIndex(mapKey)
	if !mv.IsValid() {
		return false, nil
	}

	elem := reflect.New(mv.Type()).Elem()
	elem.Set(mv)

	childRaw := rawChildRV(raw, seg)
	if indexed, indexedOK, err := tomlCaptureAOTIndexed(elem, childRaw, k, aot, idx+1, file); indexed {
		if err != nil || !indexedOK {
			return indexedOK, err
		}

		v.SetMapIndex(mapKey, elem)

		return true, nil
	}

	ok, err := tomlCaptureAt(elem, childRaw, k, aot, idx+1, file)
	if err != nil || !ok {
		return ok, err
	}

	v.SetMapIndex(mapKey, elem)

	return true, nil
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

	if elemType.Kind() == reflect.Interface && elemType.NumMethod() == 0 {
		extras.SetMapIndex(key, val)
		return nil
	}

	if val.Type().AssignableTo(elemType) {
		extras.SetMapIndex(key, val)
		return nil
	}

	if tomlValueConvertible(val, elemType) {
		extras.SetMapIndex(key, val.Convert(elemType))
		return nil
	}

	return typeMismatchRefusal(file, 0, ObjectKey(path), kindOf(elemType), tomlRawValueTerm(val))
}
