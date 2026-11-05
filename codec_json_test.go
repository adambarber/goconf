package goconf

import (
	"reflect"
	"strings"
	"testing"
)

type region string

type host string

type nestedCfg struct {
	Host host `json:"host"`
	Port int  `json:"port"`
}

type jsonCfg struct {
	Region  region    `json:"region"`
	Nested  nestedCfg `json:"nested"`
	Retries int       `json:"retries"`
}

// refusalNeed is the test's named kind for an expected refusal fragment.
type refusalNeed string

// caseName is the name of one test case.
type caseName string

// jsonCase is one strict-decode behavior: bytes in, value or refusal out.
type jsonCase struct {
	data    SourceBytes
	name    caseName
	wantErr refusalNeed
	want    jsonCfg
}

// none marks a case that must decode without refusal.
const none refusalNeed = ""

func TestDecodeJSONStrict(t *testing.T) {
	cases := []jsonCase{
		{
			name:    "valid config decodes",
			data:    SourceBytes(`{"region":"eu","retries":3,"nested":{"host":"h","port":1}}`),
			want:    jsonCfg{Region: "eu", Retries: 3, Nested: nestedCfg{Host: "h", Port: 1}},
			wantErr: none,
		},
		{
			name:    "unknown field is refused naming the field",
			data:    SourceBytes(`{"region":"eu","extra":1}`),
			want:    jsonCfg{Region: "", Retries: 0, Nested: nestedCfg{Host: "", Port: 0}},
			wantErr: refusalNeed(`unknown field "extra"`),
		},
		{
			name:    "trailing data is refused",
			data:    SourceBytes(`{"region":"eu"} {"region":"us"}`),
			want:    jsonCfg{Region: "", Retries: 0, Nested: nestedCfg{Host: "", Port: 0}},
			wantErr: refusalNeed("trailing data"),
		},
		{
			name:    "syntax error names the line",
			data:    SourceBytes("{\n  \"region\": eu\n}"),
			want:    jsonCfg{Region: "", Retries: 0, Nested: nestedCfg{Host: "", Port: 0}},
			wantErr: refusalNeed("line 2:"),
		},
		{
			name:    "type error names the field and line",
			data:    SourceBytes("{\n  \"retries\": \"many\"\n}"),
			want:    jsonCfg{Region: "", Retries: 0, Nested: nestedCfg{Host: "", Port: 0}},
			wantErr: refusalNeed("line 2:"),
		},
	}
	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			var got jsonCfg

			err := decodeJSONStrict(tc.data, Target{Value: reflect.ValueOf(&got).Elem()})
			assertJSONCase(t, &tc, got, err)
		})
	}
}

// assertJSONCase checks one decode case's refusal or value against want.
func assertJSONCase(t *testing.T, tc *jsonCase, got jsonCfg, err error) {
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

func TestPresentJSONKeys(t *testing.T) {
	data := SourceBytes(`{"region":"eu","nested":{"host":"h"},"list":[1,2]}`)

	keys, err := presentJSONKeys(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []FieldPath{"region", "nested", "nested.host", "list"} {
		if !keys.Has(want) {
			t.Errorf("expected key %q in %v", want, keys)
		}
	}

	if keys.Has("nested.port") {
		t.Errorf("nested.port is absent and must not be present")
	}
}
