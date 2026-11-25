package goconf

import "testing"

type (
	overflowInt int8
	overflowDoc struct {
		Extra map[extraKey]extraVal `goconf:"extras" json:"-"`
		N     overflowInt           `json:"n"`
	}
)

func TestExtrasJSONIntOverflowRefused(t *testing.T) {
	refuseBytesReason[overflowDoc](t, SourceBytes(`{"n":99999,"x":"y"}`), FormatJSON, ErrTypeMismatch)
}

type fixedArrDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-"`
	Arr   [2]int                `json:"arr"`
}

func TestExtrasJSONFixedArrayOverflowRefused(t *testing.T) {
	refuseBytes[fixedArrDoc](t, SourceBytes(`{"arr":[1,2,3],"x":"y"}`), FormatJSON, WantText("expected refusal"))
}

func TestExtrasJSONRootShapeRefusedAsNotAMapping(t *testing.T) {
	refuseBytesReason[resourceDoc](t, SourceBytes(`[1,2,3]`), FormatJSON, ErrNotAMapping)
}

type yamlIntExtraDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" yaml:"-"`
	N     int                   `yaml:"n"`
}

func TestExtrasYAMLTypeMismatchRefused(t *testing.T) {
	refuseBytesReason[yamlIntExtraDoc](t, SourceBytes("n: hello\nx: y\n"), FormatYAML, ErrTypeMismatch)
}

type tomlTypedExtraDoc struct {
	Resource struct {
		Extra map[extraKey]extraVal `goconf:"extras" toml:"-"`
		Name  resName               `toml:"name"`
		Kind  resKind               `toml:"kind"`
	} `toml:"resource"`
}

func TestExtrasTOMLTypedValueMismatchRefused(t *testing.T) {
	refuseBytesReason[tomlTypedExtraDoc](t, SourceBytes("[resource]\nname = \"r\"\nkind = \"k\"\n[resource.provider]\nregion = \"us\"\n"), FormatTOML, ErrTypeMismatch)
}

func TestExtrasYAMLDuplicateKeyRefused(t *testing.T) {
	refuseBytesReason[yamlIntExtraDoc](t, SourceBytes("n: 1\nx: a\nx: b\n"), FormatYAML, ErrDuplicateKey)
}

type float32ExtraDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-"`
	F     float32               `json:"f"`
}

func TestExtrasJSONFloat32OverflowRefused(t *testing.T) {
	refuseBytesReason[float32ExtraDoc](t, SourceBytes(`{"f":1e39,"x":"y"}`), FormatJSON, ErrTypeMismatch)
}

func TestExtrasTOMLIntToStringExtraRefused(t *testing.T) {
	refuseBytesReason[resourceDoc](t, SourceBytes("[resource]\nname = \"r\"\nkind = \"k\"\nextra_num = 65\n"), FormatTOML, ErrTypeMismatch)
}

func TestExtrasYAMLRootSequenceNotAMapping(t *testing.T) {
	refuseBytesReason[yamlIntExtraDoc](t, SourceBytes("- a\n- b\n"), FormatYAML, ErrNotAMapping)
}
