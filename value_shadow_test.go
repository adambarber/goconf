package goconf

import (
	"strings"
	"testing"
)

type valueArrayDoc struct {
	Array [2]Value `toml:"array"`
}

type valueSliceDoc struct {
	Slice []Value `toml:"slice"`
}

type valuePtrDoc struct {
	Ptr *Value `toml:"ptr"`
}

type valueMapDoc struct {
	Map map[valueKey]Value `toml:"map"`
}

func assertValueText(t *testing.T, v *Value, want valueWord) {
	t.Helper()

	got, err := DecodeValue[valueWord](*v)
	if err != nil {
		t.Fatalf("DecodeValue: %v", err)
	}

	if got != want {
		t.Fatalf("value = %q, want %q", got, want)
	}
}

// TestValueShadowCopyFillsEveryShape drives opaque Values through pointer,
// slice, array, and map targets so the shadow copy keeps every occurrence.
func TestValueShadowCopyFillsEveryShape(t *testing.T) {
	t.Run("pointer", func(t *testing.T) {
		cfg, _, err := LoadBytes[valuePtrDoc](SourceBytes("ptr = \"p\"\n"), Spec{Name: configName("app"), Format: FormatTOML})
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		if cfg.Ptr == nil {
			t.Fatal("pointer Value was not filled")
		}

		assertValueText(t, cfg.Ptr, "p")
	})

	t.Run("slice", func(t *testing.T) {
		cfg, _, err := LoadBytes[valueSliceDoc](SourceBytes("slice = [\"s1\", \"s2\"]\n"), Spec{Name: configName("app"), Format: FormatTOML})
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		assertValueText(t, &cfg.Slice[0], "s1")
		assertValueText(t, &cfg.Slice[1], "s2")
	})

	t.Run("array", func(t *testing.T) {
		cfg, _, err := LoadBytes[valueArrayDoc](SourceBytes("array = [\"a1\", \"a2\"]\n"), Spec{Name: configName("app"), Format: FormatTOML})
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		assertValueText(t, &cfg.Array[0], "a1")
		assertValueText(t, &cfg.Array[1], "a2")
	})

	t.Run("map", func(t *testing.T) {
		cfg, _, err := LoadBytes[valueMapDoc](SourceBytes("[map]\nk = \"m\"\n"), Spec{Name: configName("app"), Format: FormatTOML})
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		mapValue := cfg.Map["k"]
		assertValueText(t, &mapValue, "m")
	})
}

type valueFuzzDoc struct {
	Value Value `toml:"value"`
}

const valueFuzzPrefix = "FUZZVALUE-"

// FuzzDecodeValue drives arbitrary TOML text into a Value and materializes it
// into scalar targets, asserting refusals never echo the held text.
func FuzzDecodeValue(f *testing.F) {
	for _, seed := range []string{"", "abc", "42", "1.5", "true", "a'b"} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		planted := valueFuzzPrefix + strings.ReplaceAll(string(data), "\x00", "x")
		src := SourceBytes("value = '" + planted + "'\n")

		cfg, _, err := LoadBytes[valueFuzzDoc](src, Spec{Name: configName("app"), Format: FormatTOML})
		if err == nil {
			if _, derr := DecodeValue[valueWord](cfg.Value); derr != nil && strings.Contains(derr.Error(), planted) {
				t.Fatalf("DecodeValue echoed the held value: %s", derr)
			}

			return
		}

		if strings.Contains(err.Error(), planted) {
			t.Fatalf("load refusal echoed the held value: %s", err)
		}
	})
}
