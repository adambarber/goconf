package goconf

import (
	"reflect"
	"strings"
	"testing"
)

type jsonCfg struct {
	Region  region    `json:"region"`
	Nested  nestedCfg `json:"nested"`
	Retries int       `json:"retries"`
}

// jsonCase is one strict-decode behavior: bytes in, value or refusal out.
type jsonCase struct {
	data    SourceBytes
	name    caseName
	wantErr refusalNeed
	want    jsonCfg
}

// jsonWant is the zero-retries value a json refusal case decodes to.
func jsonWant() jsonCfg {
	return jsonCfg{Nested: nestedCfg{Host: "", Port: 0}, Region: "eu", Retries: 0}
}

func TestDecodeJSONStrict(t *testing.T) {
	cases := []jsonCase{
		{
			name:    nameValid,
			data:    SourceBytes(`{"region":"eu","retries":3,"nested":{"host":"h","port":1}}`),
			want:    jsonCfg{Nested: nestedCfg{Host: "h", Port: 1}, Region: "eu", Retries: 3},
			wantErr: none,
		},
		{
			name:    nameUnknown,
			data:    SourceBytes(`{"region":"eu","extra":1}`),
			want:    jsonWant(),
			wantErr: refusalNeed(`unknown field "extra"`),
		},
		{
			name:    "trailing data is refused",
			data:    SourceBytes(`{"region":"eu"} {"region":"us"}`),
			want:    jsonWant(),
			wantErr: refusalNeed("trailing data"),
		},
		{
			name:    "syntax error names the line",
			data:    SourceBytes("{\n  \"region\": eu\n}"),
			want:    jsonWant(),
			wantErr: refusalNeed("line 2:"),
		},
		{
			name:    "type error names the field and line",
			data:    SourceBytes("{\n  \"retries\": \"many\"\n}"),
			want:    jsonWant(),
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

// dupCfg is the target for duplicate-key refusals: one object and one
// scalar field that a duplicate may re-declare in whole or in part.
type dupCfg struct {
	Servers []nestedCfg `json:"servers"`
	Server  nestedCfg   `json:"server"`
	Admin   bool        `json:"admin"`
	Port    int         `json:"port"`
}

// dupCase is one duplicate-key behavior: bytes in, refusal substring or
// empty for a clean decode into want.
type dupCase struct {
	data    SourceBytes
	name    caseName
	wantErr refusalNeed
	want    dupCfg
}

func dupCases() []dupCase {
	return []dupCase{
		{
			name:    "duplicate top-level scalar key is refused",
			data:    SourceBytes(`{"admin":false,"port":1,"admin":true}`),
			want:    dupCfg{Servers: nil, Server: nestedCfg{Host: "", Port: 0}, Admin: false, Port: 0},
			wantErr: refusalNeed(`duplicate key "admin"`),
		},
		{
			name:    "duplicate top-level object key is refused",
			data:    SourceBytes(`{"server":{"host":"a"},"server":{"host":"b"}}`),
			want:    dupCfg{Servers: nil, Server: nestedCfg{Host: "", Port: 0}, Admin: false, Port: 0},
			wantErr: refusalNeed(`duplicate key "server"`),
		},
		{
			name:    "duplicate object key re-declaring part of a struct is refused",
			data:    SourceBytes(`{"server":{"host":"a"},"server":{"port":2}}`),
			want:    dupCfg{Servers: nil, Server: nestedCfg{Host: "", Port: 0}, Admin: false, Port: 0},
			wantErr: refusalNeed(`duplicate key "server"`),
		},
		{
			name:    "duplicate nested key is refused with the key named",
			data:    SourceBytes(`{"server":{"host":"a","host":"b"}}`),
			want:    dupCfg{Servers: nil, Server: nestedCfg{Host: "", Port: 0}, Admin: false, Port: 0},
			wantErr: refusalNeed(`duplicate key "host"`),
		},
		{
			name:    caseName("no duplicate keys still decodes"),
			data:    SourceBytes(`{"admin":true,"port":1,"server":{"host":"h","port":2}}`),
			want:    dupCfg{Server: nestedCfg{Host: "h", Port: 2}, Servers: nil, Admin: true, Port: 1},
			wantErr: "",
		},
		{
			name:    "repeated field names across different array elements are legal",
			data:    SourceBytes(`{"servers":[{"host":"a"},{"host":"b"}]}`),
			want:    dupCfg{Servers: []nestedCfg{{Host: "a", Port: 0}, {Host: "b", Port: 0}}, Server: nestedCfg{Host: "", Port: 0}, Admin: false, Port: 0},
			wantErr: "",
		},
	}
}

func TestDecodeJSONStrictRefusesDuplicateKeys(t *testing.T) {
	for _, tc := range dupCases() {
		t.Run(string(tc.name), func(t *testing.T) {
			var got dupCfg

			err := decodeJSONStrict(tc.data, Target{Value: reflect.ValueOf(&got).Elem()})
			assertDupCase(t, &tc, got, err)
		})
	}
}

// assertDupCase pins the duplicate-key contract: a refusal case errors
// with the key named, a clean case decodes to the exact expected value.
func assertDupCase(t *testing.T, tc *dupCase, got dupCfg, err error) {
	t.Helper()

	switch {
	case tc.wantErr == "" && err != nil:
		t.Fatalf("unexpected error: %v", err)
	case tc.wantErr != "" && err == nil:
		t.Fatalf("expected error containing %q, got nil", tc.wantErr)
	case tc.wantErr != "" && !strings.Contains(err.Error(), string(tc.wantErr)):
		t.Fatalf("error %q does not contain %q", err, tc.wantErr)
	}

	if tc.wantErr == "" && !reflect.DeepEqual(got, tc.want) {
		t.Fatalf("got %+v, want %+v", got, tc.want)
	}
}

func TestPresentJSONKeys(t *testing.T) {
	data := SourceBytes(`{"region":"eu","nested":{"host":"h"},"list":[1,2]}`)

	keys, err := presentJSONKeys(data)
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
