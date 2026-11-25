package goconf

import (
	"errors"
	"testing"
)

func TestExtrasTypedIntTOML(t *testing.T) {
	got := loadBytesOK[intExtrasDoc](t, SourceBytes("name = \"r\"\nn = 42\n"), FormatTOML)
	if got.Extra["n"] != 42 {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedBoolTOML(t *testing.T) {
	got := loadBytesOK[boolExtrasDoc](t, SourceBytes("name = \"r\"\nflag = true\n"), FormatTOML)
	if !bool(got.Extra["flag"]) {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedFloatTOML(t *testing.T) {
	got := loadBytesOK[floatExtrasDoc](t, SourceBytes("name = \"r\"\nf = 1.5\n"), FormatTOML)
	if got.Extra["f"] != 1.5 {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedNarrowIntTOML(t *testing.T) {
	got := loadBytesOK[int8ExtrasDoc](t, SourceBytes("name = \"r\"\nn = 42\n"), FormatTOML)
	if got.Extra["n"] != 42 {
		t.Fatalf("got=%v", got)
	}
}

func TestExtrasTypedUintTOMLRefused(t *testing.T) {
	err := refuseBytes[uintExtrasDoc](t, SourceBytes("name = \"r\"\nu = 7\n"), FormatTOML, WantText("expected refusal"))

	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("err=%v", err)
	}
}

func TestExtrasTypedInt8OverflowTOML(t *testing.T) {
	err := refuseBytes[int8ExtrasDoc](t, SourceBytes("name = \"r\"\nn = 999999\n"), FormatTOML, WantText("expected refusal"))

	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("err=%v", err)
	}
}
