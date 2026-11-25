package goconf

import (
	"encoding"
	"encoding/json"
	"reflect"
	"strconv"
)

// envParsed is the set of scalar shapes an env value can become.
type envParsed interface {
	bool | int64 | uint64 | float64
}

func setEnvValue(name EnvName, dst reflect.Value, raw envRawValue) error {
	t := dst.Type()

	switch t.Kind() {
	case reflect.String:
		dst.SetString(string(raw))
	case reflect.Bool:
		return setEnvBool(name, dst, raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setEnvInt(name, dst, raw)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setEnvUint(name, dst, raw)
	case reflect.Float32, reflect.Float64:
		return setEnvFloat(name, dst, raw)
	case reflect.Slice, reflect.Map, reflect.Struct, reflect.Array:
		return parseEnvComposite(name, dst, raw)
	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Pointer, reflect.UnsafePointer:
		return envParseRefusal(name, t)
	default:
		return envParseRefusal(name, t)
	}

	return nil
}

// parseEnv folds one parse and one setter into the single refusal branch.
func parseEnv[T envParsed](name EnvName, dst reflect.Value, raw envRawValue,
	parse func(envRawValue) (T, bool), set func(reflect.Value, T),
) error {
	parsed, ok := parse(raw)
	if !ok {
		return envParseRefusal(name, dst.Type())
	}

	set(dst, parsed)

	return nil
}

func setEnvBool(name EnvName, dst reflect.Value, raw envRawValue) error {
	return parseEnv(name, dst, raw, parseBoolValue(), setBoolValue)
}

func setEnvInt(name EnvName, dst reflect.Value, raw envRawValue) error {
	parse := func(raw envRawValue, bits envBitSize) (int64, error) {
		return strconv.ParseInt(string(raw), 10, int(bits))
	}

	return parseEnv(name, dst, raw, envParseFactory(dst, parse), setIntValue)
}

func setEnvUint(name EnvName, dst reflect.Value, raw envRawValue) error {
	parse := func(raw envRawValue, bits envBitSize) (uint64, error) {
		return strconv.ParseUint(string(raw), 10, int(bits))
	}

	return parseEnv(name, dst, raw, envParseFactory(dst, parse), setUintValue)
}

func setEnvFloat(name EnvName, dst reflect.Value, raw envRawValue) error {
	parse := func(raw envRawValue, bits envBitSize) (float64, error) {
		return strconv.ParseFloat(string(raw), int(bits))
	}

	return parseEnv(name, dst, raw, envParseFactory(dst, parse), setFloatValue)
}

func parseBoolValue() func(envRawValue) (bool, bool) {
	return func(raw envRawValue) (bool, bool) {
		parsed, err := strconv.ParseBool(string(raw))

		return parsed, err == nil
	}
}

// envBitSize is the declared width of a numeric field.
type envBitSize int

// envParseFactory turns one width-aware parser into a value parser, so the
// integer, unsigned, and float paths share a single refusal branch.
func envParseFactory[T envParsed](dst reflect.Value, parse func(envRawValue, envBitSize) (T, error)) func(envRawValue) (T, bool) {
	bits := envBitSize(dst.Type().Bits())

	return func(raw envRawValue) (T, bool) {
		parsed, err := parse(raw, bits)

		return parsed, err == nil
	}
}

func setBoolValue(dst reflect.Value, parsed bool)   { dst.SetBool(parsed) }
func setIntValue(dst reflect.Value, parsed int64)   { dst.SetInt(parsed) }
func setUintValue(dst reflect.Value, parsed uint64) { dst.SetUint(parsed) }
func setFloatValue(dst reflect.Value, parsed float64) {
	dst.SetFloat(parsed)
}

func parseEnvComposite(name EnvName, dst reflect.Value, raw envRawValue) error {
	t := dst.Type()

	if decoder, ok := reflect.TypeAssert[encoding.TextUnmarshaler](dst.Addr()); ok {
		if decoder.UnmarshalText([]byte(raw)) == nil {
			return nil
		}

		return envParseRefusal(name, t)
	}

	if err := json.Unmarshal([]byte(raw), dst.Addr().Interface()); err != nil {
		return envParseRefusal(name, t)
	}

	return nil
}
