package goconf

import (
	"errors"
	"testing"
)

// ConfigTarget is the closed set of config structs the shared load helpers fill.
type ConfigTarget interface {
	loadCfg | resourceDoc | envCfg | nestedDynCfg | dynExtrasDoc | yamlMapDoc |
		jsonUnknownExtrasDoc | intExtrasDoc | int8ExtrasDoc | uintExtrasDoc |
		floatExtrasDoc | boolExtrasDoc | sliceExtrasDoc | structExtrasDoc |
		float32ExtrasDoc | mapExtrasDoc | ptrExtrasDoc | mapFieldDoc |
		overflowDoc | fixedArrDoc | yamlIntExtraDoc | tomlTypedExtraDoc | float32ExtraDoc |
		metaItemsCfg | metaNumsCfg | metaMergeCfg | metaMergeSeqCfg | metaNameOtherCfg |
		metaDynamicJSONCfg | metaExtrasTOMLCfg | metaRegionRetriesCfg | metaTOMLItemsCfg | mrcfg | muCfg
}

// jsonUnknownExtrasDoc is the shared target for nested unknown-field tests.
type jsonUnknownExtrasDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-"`
	In    struct {
		Name resName `json:"name"`
	} `json:"in"`
}

// yamlMapDoc is the shared target for YAML map-shape mismatch tests.
type yamlMapDoc struct {
	Extra map[extraKey]extraVal `goconf:"extras" yaml:"-"`
	M     map[extraKey]extraVal `yaml:"m"`
}

// loadBytes runs one in-memory source through the public API.
func loadBytes[T ConfigTarget](t *testing.T, data SourceBytes, format Format) (T, error) {
	t.Helper()

	got, _, err := LoadBytes[T]([]byte(data), Spec{Name: configName("app"), Format: format})

	return got, err
}

// loadBytesOK loads one in-memory source and fails on any error.
func loadBytesOK[T ConfigTarget](t *testing.T, data SourceBytes, format Format) T {
	t.Helper()

	got, err := loadBytes[T](t, data, format)
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}

	return got
}

// refuseBytes loads one in-memory source and fails unless it is refused.
func refuseBytes[T ConfigTarget](t *testing.T, data SourceBytes, format Format, want WantText) {
	t.Helper()

	if err := refuseBytesErr[T](t, data, format); err == nil {
		t.Fatalf("%s", want)
	}
}

// refuseBytesErr is refuseBytes with the refusal returned for reason checks.
func refuseBytesErr[T ConfigTarget](t *testing.T, data SourceBytes, format Format) error {
	t.Helper()

	_, err := loadBytes[T](t, data, format)

	return err
}

// refuseBytesReason checks both that a source is refused and why.
func refuseBytesReason[T ConfigTarget](t *testing.T, data SourceBytes, format Format, want Reason) {
	t.Helper()

	err := refuseBytesErr[T](t, data, format)
	if err == nil {
		t.Fatalf("expected refusal with reason %v", want)
	}

	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}
