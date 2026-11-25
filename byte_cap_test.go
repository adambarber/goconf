package goconf

import (
	"errors"
	"strings"
	"testing"
)

type capCfg struct {
	Port int `toml:"port"`
}

func capSource() SourceBytes {
	return SourceBytes("port = 1\n# " + strings.Repeat("x", 64))
}

func TestMaxBytesRefusesOversizedFile(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{fileNameTOML: capSource()})

	_, _, err := Load[capCfg](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(legPaths(root, dirLocal)...), WithMaxBytes(8))
	if !errors.Is(err, ErrSourceTooLarge) {
		t.Fatalf("err = %v, want ErrSourceTooLarge", err)
	}
}

func TestMaxBytesRefusesOversizedBytes(t *testing.T) {
	_, _, err := LoadBytes[capCfg](capSource(), Spec{Name: configName("app"), Format: FormatTOML}, WithMaxBytes(8))
	if !errors.Is(err, ErrSourceTooLarge) {
		t.Fatalf("err = %v, want ErrSourceTooLarge", err)
	}
}

func TestMaxBytesDefaultAdmitsOrdinaryFile(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{fileNameTOML: capSource()})

	got, _, err := Load[capCfg](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(legPaths(root, dirLocal)...))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.Port != 1 {
		t.Fatalf("port = %d, want 1", got.Port)
	}
}

func TestMaxBytesDefaultLimitApplies(t *testing.T) {
	big := make([]byte, int(defaultByteCap)+1)

	_, _, err := LoadBytes[capCfg](big, Spec{Name: configName("app"), Format: FormatTOML})
	if !errors.Is(err, ErrSourceTooLarge) {
		t.Fatalf("err = %v, want ErrSourceTooLarge", err)
	}
}

func TestMaxBytesCapIsInclusive(t *testing.T) {
	src := SourceBytes("port=1\n\n")

	got, _, err := LoadBytes[capCfg](src, Spec{Name: configName("app"), Format: FormatTOML}, WithMaxBytes(8))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.Port != 1 {
		t.Fatalf("port = %d, want 1", got.Port)
	}
}

func TestMaxBytesPanicsOutsideRange(t *testing.T) {
	for _, n := range []int{0, -1, int(hardByteCap) + 1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("WithMaxBytes(%d) must panic", n)
				}
			}()

			WithMaxBytes(n)
		}()
	}
}

func TestMaxBytesRefusalRendersOneAction(t *testing.T) {
	_, _, err := LoadBytes[capCfg](capSource(), Spec{Name: configName("app"), Format: FormatTOML}, WithMaxBytes(8))
	if err == nil {
		t.Fatal("oversized source must be refused")
	}

	want := "Error: app:1: source too large: the config file. split it into smaller files, or raise the size limit for this load."
	if err.Error() != want {
		t.Fatalf("render = %q, want %q", err.Error(), want)
	}
}
