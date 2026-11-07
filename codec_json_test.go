package goconf

import (
	"reflect"
	"testing"
)

func TestDecodeJSONStrict(t *testing.T) {
	cases := []codecCase{
		{
			name:    nameValid,
			data:    SourceBytes(`{"region":"eu","retries":3,"nested":{"host":"h","port":1}}`),
			want:    codecWant(3, nestedCfgWant("h", 1)),
			wantErr: none,
		},
		{
			name:    nameUnknown,
			data:    SourceBytes(`{"region":"eu","extra":1}`),
			want:    zeroCodecWant(),
			wantErr: refusalNeed(`unknown field "extra"`),
		},
		{
			name:    "unknown field names the key's line, no doubled prefix",
			data:    SourceBytes("{\n  \"region\": \"eu\",\n  \"extra\": 1\n}"),
			want:    zeroCodecWant(),
			wantErr: refusalNeed(`line 3: unknown field "extra"`),
		},
		{
			name:    "trailing data is refused",
			data:    SourceBytes(`{"region":"eu"} {"region":"us"}`),
			want:    zeroCodecWant(),
			wantErr: refusalNeed("trailing data"),
		},
		{
			name:    "syntax error names the line",
			data:    SourceBytes("{\n  \"region\": eu\n}"),
			want:    zeroCodecWant(),
			wantErr: refusalNeed("line 2:"),
		},
		{
			name:    "type error names the field and line",
			data:    SourceBytes("{\n  \"retries\": \"many\"\n}"),
			want:    zeroCodecWant(),
			wantErr: refusalNeed("line 2:"),
		},
	}
	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			decodeCase(t, &tc, decodeJSONStrict)
		})
	}
}

// dupCase is one duplicate-key behavior: bytes in, refusal substring or
// empty for a clean decode into want.
type dupCase struct {
	data    SourceBytes
	name    caseName
	wantErr refusalNeed
	want    dupCfg
}

// dupCfg is the target for duplicate-key refusals: one object and one
// scalar field that a duplicate may re-declare in whole or in part.
type dupCfg struct {
	Servers []nestedCfg `json:"servers"`
	Server  nestedCfg   `json:"server"`
	Admin   bool        `json:"admin"`
	Port    int         `json:"port"`
}

func dupCases() []dupCase {
	zero := dupCfg{Server: nestedCfgWant("", 0), Servers: nil, Admin: false, Port: 0}

	return []dupCase{
		{
			name:    "duplicate top-level scalar key is refused",
			data:    SourceBytes(`{"admin":false,"port":1,"admin":true}`),
			want:    zero,
			wantErr: refusalNeed(`duplicate key "admin"`),
		},
		{
			name:    "duplicate top-level object key is refused",
			data:    SourceBytes(`{"server":{"host":"a"},"server":{"host":"b"}}`),
			want:    zero,
			wantErr: refusalNeed(`duplicate key "server"`),
		},
		{
			name:    "duplicate object key re-declaring part of a struct is refused",
			data:    SourceBytes(`{"server":{"host":"a"},"server":{"port":2}}`),
			want:    zero,
			wantErr: refusalNeed(`duplicate key "server"`),
		},
		{
			name:    "duplicate nested key is refused with the key named",
			data:    SourceBytes(`{"server":{"host":"a","host":"b"}}`),
			want:    zero,
			wantErr: refusalNeed(`duplicate key "host"`),
		},
		{
			name:    "duplicate key after an array close is refused",
			data:    SourceBytes(`{"servers":[{"host":"a"}],"servers":[{"host":"b"}]}`),
			want:    zero,
			wantErr: refusalNeed(`duplicate key "servers"`),
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

	if refusalSatisfied(t, tc.wantErr, err) {
		return
	}

	if !reflect.DeepEqual(got, tc.want) {
		t.Fatalf("got %+v, want %+v", got, tc.want)
	}
}

func TestPresentJSONKeys(t *testing.T) {
	data := SourceBytes(`{"region":"eu","nested":{"host":"h"},"list":[1,2]}`)

	keys := presentKeysOK(t, presentJSONKeys, data)

	keysPresent(t, keys, []FieldPath{fpRegion, fpNested, fpNestedHost, fpList}, []FieldPath{fpNestedPort})
}
