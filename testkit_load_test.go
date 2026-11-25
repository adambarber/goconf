package goconf

import "testing"

// ConfigTarget is the closed set of config structs the shared load helpers fill.
type ConfigTarget interface {
	loadCfg | resourceDoc | envCfg | nestedDynCfg | yamlMapDoc |
		intExtrasDoc | int8ExtrasDoc | uintExtrasDoc | floatExtrasDoc | boolExtrasDoc
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
