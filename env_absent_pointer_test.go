package goconf

import "testing"

type envAbsentHost string

type envAbsentInner struct {
	Host envAbsentHost `json:"host"`
}

type envAbsentCfg struct {
	TLS *envAbsentInner `json:"tls"`
}

func TestWithEnvLeavesUnsetPointerFieldNil(t *testing.T) {
	got, _, err := LoadBytes[envAbsentCfg](SourceBytes(`{}`),
		Spec{Name: configName("app"), Format: FormatJSON},
		WithEnv(EnvPrefix("GOCONF_REVIEW_ABSENT_POINTER")))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.TLS != nil {
		t.Fatalf("TLS = %+v, want nil with no matching variable", got.TLS)
	}
}
