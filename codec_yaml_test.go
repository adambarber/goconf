package goconf

import (
	"reflect"
	"strings"
	"testing"
)

type yamlCfg struct {
	Region  region    `json:"region"`
	Nested  nestedCfg `json:"nested"`
	Retries int       `json:"retries"`
}

// yamlCase is one strict-decode behavior: bytes in, value or refusal out.
type yamlCase struct {
	data    SourceBytes
	name    caseName
	wantErr refusalNeed
	want    yamlCfg
}

// yamlWant is the zero-retries value the null and type-error cases yield.
func yamlWant() yamlCfg {
	return yamlCfg{Nested: nestedCfg{Host: "", Port: 0}, Region: "eu", Retries: 0}
}

func TestDecodeYAMLStrict(t *testing.T) {
	cases := []yamlCase{
		{
			name:    nameValid,
			data:    SourceBytes("region: eu\nretries: 3\nnested:\n  host: h\n  port: 1\n"),
			want:    yamlCfg{Nested: nestedCfg{Host: "h", Port: 1}, Region: "eu", Retries: 3},
			wantErr: none,
		},
		{
			name:    nameUnknown,
			data:    SourceBytes("region: eu\nextra: 1\n"),
			want:    yamlWant(),
			wantErr: refusalNeed(`field extra not found`),
		},
		{
			name:    "explicit null is accepted and left zero",
			data:    SourceBytes("region: eu\nretries: null\n"),
			want:    yamlWant(),
			wantErr: none,
		},
		{
			name:    "type error names the line",
			data:    SourceBytes("region: eu\nretries: many\n"),
			want:    yamlWant(),
			wantErr: refusalNeed("line 2:"),
		},
	}
	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			var got yamlCfg

			err := decodeYAMLStrict(tc.data, Target{Value: reflect.ValueOf(&got).Elem()})
			assertYAMLCase(t, &tc, got, err)
		})
	}
}

func assertYAMLCase(t *testing.T, tc *yamlCase, got yamlCfg, err error) {
	t.Helper()

	switch {
	case tc.wantErr == "" && err != nil:
		t.Fatalf("unexpected error: %v", err)
	case tc.wantErr != "" && err == nil:
		t.Fatalf("expected error containing %q, got nil", tc.wantErr)
	case tc.wantErr != "" && !strings.Contains(err.Error(), string(tc.wantErr)):
		t.Fatalf("error %q does not contain %q", err, tc.wantErr)
	}

	if tc.wantErr == "" && got != tc.want {
		t.Fatalf("got %+v, want %+v", got, tc.want)
	}
}

func TestPresentYAMLKeys(t *testing.T) {
	data := SourceBytes("region: eu\nnested:\n  host: h\nlist:\n  - a\n  - b\n")

	keys, err := presentYAMLKeys(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []FieldPath{fpRegion, fpNested, fpNestedHost, fpList} {
		if !keys.Has(want) {
			t.Errorf("expected key %q in %v", want, keys)
		}
	}

	if keys.Has(fpNestedPort) {
		t.Errorf("nested.port is absent and must not be present")
	}
}
