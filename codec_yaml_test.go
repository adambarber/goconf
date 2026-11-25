package goconf

import "testing"

func TestDecodeYAMLStrict(t *testing.T) {
	cases := []codecCase{
		{
			name:    nameValid,
			data:    SourceBytes("region: eu\nretries: 3\nnested:\n  host: h\n  port: 1\n"),
			want:    codecWant(3, nestedCfgWant("h", 1)),
			wantErr: none,
		},
		{
			name:    nameUnknown,
			data:    SourceBytes("region: eu\nextra: 1\n"),
			want:    zeroCodecWant(),
			wantErr: refusalNeed(`unknown field: "extra"`),
		},
		{
			name:    "explicit null is accepted and left zero",
			data:    SourceBytes("region: eu\nretries: null\n"),
			want:    zeroCodecWant(),
			wantErr: none,
		},
		{
			name:    "type error names the line",
			data:    SourceBytes("region: eu\nretries: many\n"),
			want:    zeroCodecWant(),
			wantErr: refusalNeed(`the value of "retries" is text, not a whole number`),
		},
	}
	for _, tc := range cases {
		t.Run(string(tc.name), func(t *testing.T) {
			decodeCase(t, &tc, fileDecode(decodeYAMLStrict, fileNameYAML))
		})
	}
}

// yamlPresenceCfg is the public-API target for the YAML presence walk test.
type yamlPresenceCfg struct {
	Region mstr      `yaml:"region"`
	Nested nestedCfg `yaml:"nested"`
	List   []mstr    `yaml:"list"`
}

// TestLoadWithMetaYAMLPresence pins the YAML presence walk through LoadWithMeta:
// keys inside mappings and sequences are recorded under their schema paths.
func TestLoadWithMetaYAMLPresence(t *testing.T) {
	_, _, meta, err := LoadBytesWithMeta[yamlPresenceCfg]([]byte("region: eu\nnested:\n  host: h\nlist:\n  - a\n  - b\n"),
		Spec{Name: configName("app"), Format: FormatYAML})
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	assertSchemaPaths(t, meta, spRegion, spNested, spNestedHost, "list")
	assertSchemaAbsent(t, meta, "nested.port")
}
