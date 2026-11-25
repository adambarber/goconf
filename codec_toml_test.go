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
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("region = \"eu\"\n\n[nested]\nhost = \"h\"\n")},
	})

	_, _, meta, err := LoadWithMeta[codecCfg](Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	for _, p := range []SchemaPath{spRegion, spNested, spNestedHost} {
		if len(meta.BySchemaPath(p)) == 0 {
			t.Errorf("schema path %q missing: %#v", p, meta.Entries())
		}
	}

	if len(meta.BySchemaPath("nested.port")) != 0 {
		t.Errorf("nested.port is absent and must not be present")
	}
}
