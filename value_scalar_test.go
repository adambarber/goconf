package goconf

import (
	"errors"
	"testing"
)

type valueScalarCfg struct {
	Value Value `toml:"value"`
}

func loadScalarValue(t *testing.T, body SourceBytes) Value {
	t.Helper()

	got, _, err := LoadBytes[valueScalarCfg](body, Spec{Name: configName("app"), Format: FormatTOML})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	return got.Value
}

func TestDecodeValueMaterializesTOMLScalars(t *testing.T) {
	cases := map[string]struct {
		want valueKind
		body SourceBytes
	}{
		"integer": {want: kindInt, body: SourceBytes("value = 42\n")},
		"float":   {want: kindFloat, body: SourceBytes("value = 1.5\n")},
		"boolean": {want: kindBool, body: SourceBytes("value = true\n")},
		"string":  {want: kindString, body: SourceBytes("value = \"hello\"\n")},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			value := loadScalarValue(t, tc.body)

			switch tc.want {
			case kindInt:
				assertDecodedValue(t, &value, 42)
			case kindFloat:
				assertDecodedValue(t, &value, 1.5)
			case kindBool:
				assertDecodedValue(t, &value, true)
			case kindString:
				assertDecodedValue(t, &value, "hello")
			case kindList, kindTable:
				t.Fatalf("unexpected scalar kind %q", tc.want)
			}
		})
	}
}

func assertDecodedValue[T comparable](t *testing.T, v *Value, want T) {
	t.Helper()

	got, err := DecodeValue[T](*v)
	if err != nil {
		t.Fatalf("DecodeValue: %v", err)
	}

	if got != want {
		t.Fatalf("DecodeValue = %v, want %v", got, want)
	}
}

func TestDecodeValueMaterializesPointerScalar(t *testing.T) {
	value, err := DecodeValue[*int](loadScalarValue(t, SourceBytes("value = 42\n")))
	if err != nil {
		t.Fatalf("DecodeValue[*int]: %v", err)
	}

	if value == nil || *value != 42 {
		t.Fatalf("DecodeValue[*int] = %v, want 42", value)
	}
}

func TestDecodeValueRefusesWrongScalarType(t *testing.T) {
	_, err := DecodeValue[string](loadScalarValue(t, SourceBytes("value = 42\n")))
	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("err = %v, want ErrTypeMismatch", err)
	}
}
