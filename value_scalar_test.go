package goconf

import (
	"errors"
	"testing"
)

type valueScalarCfg struct {
	Value Value `toml:"value"`
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
			got, _, err := LoadBytes[valueScalarCfg](tc.body,
				Spec{Name: configName("app"), Format: FormatTOML})
			if err != nil {
				t.Fatalf("load: %v", err)
			}

			switch tc.want {
			case kindInt:
				assertDecodedValue(t, &got.Value, 42)
			case kindFloat:
				assertDecodedValue(t, &got.Value, 1.5)
			case kindBool:
				assertDecodedValue(t, &got.Value, true)
			case kindString:
				assertDecodedValue(t, &got.Value, "hello")
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

func TestDecodeValueRefusesWrongScalarType(t *testing.T) {
	got, _, err := LoadBytes[valueScalarCfg](SourceBytes("value = 42\n"),
		Spec{Name: configName("app"), Format: FormatTOML})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	_, err = DecodeValue[string](got.Value)
	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("err = %v, want ErrTypeMismatch", err)
	}
}
