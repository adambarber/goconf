package goconf

import "testing"

const (
	metaPathItems0Name OccurrencePath = "items[0].name"
	metaPathItems1Name OccurrencePath = "items[1].name"
)

// Shared config shapes for the metadata behavior tests. One tagged struct
// serves JSON, YAML and TOML so the same source bytes exercise each collector.
type (
	metaItem struct {
		Name mstr `json:"name" toml:"name" yaml:"name"`
	}

	metaItemsCfg struct {
		Items []metaItem `json:"items" toml:"items" yaml:"items"`
	}

	metaNumsCfg struct {
		Nums []int `json:"nums" toml:"nums" yaml:"nums"`
	}

	metaInner struct {
		X int `yaml:"x"`
		Y int `yaml:"y"`
		Z int `yaml:"z"`
	}

	metaMergeCfg struct {
		Base metaInner `yaml:"base"`
		A    metaInner `yaml:"a"`
	}

	metaMergeSeqCfg struct {
		B1 metaInner `yaml:"b1"`
		B2 metaInner `yaml:"b2"`
		A  metaInner `yaml:"a"`
	}

	metaNameOtherCfg struct {
		Name  mstr `toml:"name"`
		Other mstr `toml:"other"`
	}

	metaDynamicJSONCfg struct {
		Env   map[mstr]mstr `goconf:"dynamic" json:"env"`
		Extra map[mstr]mstr `goconf:"extras"  json:"-"`
		Name  mstr          `json:"name"`
	}

	metaExtrasTOMLCfg struct {
		Extra map[mstr]mstr `goconf:"extras" toml:"-"`
		Name  mstr          `toml:"name"`
	}

	metaRegionRetriesCfg struct {
		Region  mstr `json:"region"`
		Retries int  `json:"retries"`
	}

	metaTOMLItemsCfg struct {
		Items []metaItem `toml:"items"`
	}
)

// loadMetaGlobal loads one source from the global directory and returns its
// decoded value and metadata. It is the single entry point for metadata tests.
func loadMetaGlobal[T ConfigTarget](t *testing.T, data SourceBytes, format Format, opts ...Option) (T, Meta) {
	t.Helper()

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {metaFileName(format): data},
	})
	t.Chdir(string(root))

	allOpts := append(loadOpts(globalPathOf(root)), opts...)

	got, _, meta, err := LoadWithMeta[T](Spec{Name: configName("app"), Format: format}, allOpts...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	return got, meta
}

// metaFileName maps a format to its canonical file name.
func metaFileName(format Format) FileName {
	if format == FormatJSON {
		return fileNameJSON
	}

	if format == FormatYAML {
		return fileNameYAML
	}

	return fileNameTOML
}
