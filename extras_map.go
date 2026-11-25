package goconf

import "reflect"

// setExtraMapValue is the single write path for extras maps. Callers decide
// how a value is decoded and how a non-map destination is refused.
func setExtraMapValue(extras reflect.Value, key keySegment, fill func(reflect.Value) error) error {
	if extras.IsNil() && extras.CanSet() {
		extras.Set(reflect.MakeMap(extras.Type()))
	}

	elem := reflect.New(extras.Type().Elem()).Elem()
	if err := fill(elem); err != nil {
		return err
	}

	extras.SetMapIndex(reflect.ValueOf(string(key)).Convert(extras.Type().Key()), elem)

	return nil
}
