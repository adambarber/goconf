package goconf

import (
	"reflect"
	"strings"
	"testing"
)

type tomlCfg struct {
	Region  region    `json:"region"`
	Nested  nestedCfg `json:"nested"`
	Retries int       `json:"retries"`
}

// tomlCase is one strict-decode behavior: bytes in, value or refusal out.
type tomlCase struct {
	data    SourceBytes
	name    caseName
	wantErr refusalNeed
	want    tomlCfg
}

// tomlWant is the zero value a toml refusal case decodes to.
func tomlWant() tomlCfg {
	return tomlCfg{Nested: nestedCfg{Host: "", Port: 0}, Region: "eu", Retries: 0}
}

func TestDecodeTOMLStrict(t *testing.T) {
	cases := []tomlCase{
		{
			name:    nameValid,
			data:    SourceBytes("region = \"eu\"\nretries = 3\n\n[nested]\nhost = \"h\"\nport = 1\n"),
			want:    tomlCfg{Nested: nestedCfg{Host: "h", Port: 1}, Region: "eu", Retries: 3},
			wantErr: none,
		},
		{
			name:    nameUnknown,
			data:    SourceBytes("region = \"eu\"\nextra = 1\n"),
			want:    tomlWant(),
			wantErr: refusalNeed(`unknown field: "extra"`),
		},
		{
			name:    "type error names the field",
			data:    SourceBytes("retries = \"many\"\n"),
			want:    tomlWant(),
			wantErr: refusalNeed("retries"),
		},
	}
	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			var got tomlCfg

			err := decodeTOMLStrict(tc.data, Target{Value: reflect.ValueOf(&got).Elem()})
			assertTOMLCase(t, &tc, got, err)
		})
	}
}

func assertTOMLCase(t *testing.T, tc *tomlCase, got tomlCfg, err error) {
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

func TestPresentTOMLKeys(t *testing.T) {
	data := SourceBytes("region = \"eu\"\n\n[nested]\nhost = \"h\"\n")

	keys, err := presentTOMLKeys(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []FieldPath{fpRegion, fpNested, fpNestedHost} {
		if !keys.Has(want) {
			t.Errorf("expected key %q in %v", want, keys)
		}
	}

	if keys.Has(fpNestedPort) {
		t.Errorf("nested.port is absent and must not be present")
	}
}
