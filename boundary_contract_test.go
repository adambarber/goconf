package goconf

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestInvalidFormatReturnsZeroResults pins the shared refusal shape at the
// settings boundary of both Load entry points.
func TestInvalidFormatReturnsZeroResults(t *testing.T) {
	cfg, prov, meta, err := LoadBytesWithMeta[queryCfg]([]byte(`{}`),
		Spec{Name: configName("app"), Format: Format(".yml")})
	if err == nil {
		t.Fatal("invalid format was accepted")
	}

	if cfg.Region != "" || prov != nil || meta.Len() != 0 {
		t.Fatalf("refusal leaked results: cfg=%+v prov=%v meta=%+v", cfg, prov, meta)
	}
}

type implicitScalar string

func implicitScalarNode(value implicitScalar) *yaml.Node {
	return &yaml.Node{
		Kind:        yaml.ScalarNode,
		Style:       0,
		Tag:         "",
		Value:       string(value),
		Anchor:      "",
		Alias:       nil,
		Content:     nil,
		HeadComment: "",
		LineComment: "",
		FootComment: "",
		Line:        0,
		Column:      0,
	}
}

// TestYAMLImplicitScalarFallback pins the untagged scalar resolver used when
// yaml.v3 leaves a node without a resolved core tag.
func assertYAMLNullScalars(t *testing.T) {
	t.Helper()

	for _, null := range []implicitScalar{"~", "null", "Null", "NULL", ""} {
		if got := yamlScalarImplicit(implicitScalarNode(null)); got.IsValid() {
			t.Fatalf("null value %q resolved to %v", null, got)
		}
	}
}

func assertYAMLBoolAndFloatScalars(t *testing.T) {
	t.Helper()

	if got := yamlScalarImplicit(implicitScalarNode("1.5")); !got.IsValid() || got.Float() != 1.5 {
		t.Fatalf("float resolved to %v", got)
	}

	if got := yamlScalarImplicit(implicitScalarNode("true")); !got.IsValid() || !got.Bool() {
		t.Fatalf("true resolved to %v", got)
	}
}

func TestYAMLImplicitScalarFallback(t *testing.T) {
	cases := []struct {
		value implicitScalar
		want  int64
	}{
		{"7", 7},
		{"0x10", 16},
	}

	for _, tc := range cases {
		got := yamlScalarImplicit(implicitScalarNode(tc.value))
		if !got.IsValid() || got.Kind().String() != "int64" || got.Int() != tc.want {
			t.Fatalf("value %q resolved to %v, want %d", tc.value, got, tc.want)
		}
	}

	assertYAMLNullScalars(t)
	assertYAMLBoolAndFloatScalars(t)
}
