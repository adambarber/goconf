package goconf

import (
	"reflect"
	"testing"
)

func TestDecodeTOMLStrict(t *testing.T) {
	cases := []codecCase{
		{
			name:    nameValid,
			data:    SourceBytes("region = \"eu\"\nretries = 3\n\n[nested]\nhost = \"h\"\nport = 1\n"),
			want:    codecWant(3, nestedCfgWant("h", 1)),
			wantErr: none,
		},
		{
			name:    nameUnknown,
			data:    SourceBytes("region = \"eu\"\nextra = 1\n"),
			want:    zeroCodecWant(),
			wantErr: refusalNeed(`unknown field: "extra"`),
		},
		{
			name:    "type error names the field",
			data:    SourceBytes("retries = \"many\"\n"),
			want:    zeroCodecWant(),
			wantErr: refusalNeed("retries"),
		},
	}
	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			decodeCase(t, &tc, decodeTOMLStrict)
		})
	}
}

func TestPresentTOMLKeys(t *testing.T) {
	data := SourceBytes("region = \"eu\"\n\n[nested]\nhost = \"h\"\n")

	var want codecCfg

	keys := presentKeysOK(t, func(data SourceBytes) (KeySet, error) {
		return presentTOMLKeys(data, reflect.TypeOf(want))
	}, data)

	keysPresent(t, keys, []FieldPath{fpRegion, fpNested, fpNestedHost}, []FieldPath{fpNestedPort})
}
