package goconf

import "testing"

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
			decodeCase(t, &tc, fileDecode(decodeTOMLStrict, fileNameTOML))
		})
	}
}

// TestLoadWithMetaTOMLPresence pins the TOML presence walk through the
// public surface; it is the reference implementation the other formats match.
func TestLoadWithMetaTOMLPresence(t *testing.T) {
	_, _, meta, err := LoadBytesWithMeta[codecCfg]([]byte("region = \"eu\"\n\n[nested]\nhost = \"h\"\n"),
		Spec{Name: configName("app"), Format: FormatTOML})
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	assertSchemaPaths(t, meta, spRegion, spNested, spNestedHost)
	assertSchemaAbsent(t, meta, "nested.port")
}
