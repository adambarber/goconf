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

		ok, err := tomlCaptureKey(dst.Value, raw, k)
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

func tomlCaptureKey(dst, raw reflect.Value, k toml.Key) (bool, error) {
	if len(k) == 0 {
		return false, nil
	}

	v := derefValue(dst)
	if !v.IsValid() {
		return false, nil
	}

	return tomlCaptureAt(v, raw, k, 0)
}

func tomlCaptureAt(v, raw reflect.Value, k toml.Key, idx int) (bool, error) {
	if idx >= len(k) {
		return false, nil
	}

	seg := keySegment(k[idx])
	v = derefValue(v)

	if v.Kind() == reflect.Struct {
		return tomlCaptureStructAt(v, raw, k, idx, seg)
	}

	if v.Kind() == reflect.Map {
		return tomlCaptureMapAt(v, raw, k, idx, seg)
	}

	return false, nil
}

func tomlCaptureStructAt(v, raw reflect.Value, k toml.Key, idx int, seg keySegment) (bool, error) {
	sf, ok := findStructFieldByKey(v.Type(), seg)
	if ok {
		return tomlCaptureAt(v.FieldByIndex(sf.Index), rawChildRV(raw, seg), k, idx+1)
	}

	ef, ok := extrasField(v.Type())
	if !ok {
		return false, nil
	}

	return tomlSetExtra(v.FieldByIndex(ef.Index), raw, k, idx)
}

func tomlCaptureMapAt(v, raw reflect.Value, k toml.Key, idx int, seg keySegment) (bool, error) {
	mv := mapGetStringKey(v, seg)
	if !mv.IsValid() {
		return false, nil
	}

	return tomlCaptureAt(mv, rawChildRV(raw, seg), k, idx+1)
}

func tomlSetExtra(extras, raw reflect.Value, k toml.Key, idx int) (bool, error) {
	if extras.Kind() != reflect.Map {
		return false, nil
	}

	if extras.IsNil() && extras.CanSet() {
		extras.Set(reflect.MakeMap(extras.Type()))
	}

	key := reflect.ValueOf(k[idx]).Convert(extras.Type().Key())
	val := rawChildRV(raw, keySegment(k[idx]))
	setExtraValue(extras, key, val)

	return true, nil
}

func setExtraValue(extras, key, val reflect.Value) {
	elemType := extras.Type().Elem()

	if !val.IsValid() {
		extras.SetMapIndex(key, reflect.Zero(elemType))
		return
	}

	if val.Type().ConvertibleTo(elemType) {
		extras.SetMapIndex(key, val.Convert(elemType))
		return
	}

	if elemType.Kind() == reflect.Interface && elemType.NumMethod() == 0 {
		extras.SetMapIndex(key, val)
		return
	}

	extras.SetMapIndex(key, val)
}

func decodeTOMLRaw(data SourceBytes, file FilePath) (reflect.Value, error) {
	rv := reflect.New(reflect.TypeOf((*interface{})(nil)).Elem())
	if _, err := toml.Decode(string(data), rv.Interface()); err != nil {
		return reflect.Value{}, tomlDecodeRefusal(file, err)
	}

	return rv.Elem(), nil
}
