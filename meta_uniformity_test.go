package goconf

import (
	"testing"
)

type muItem struct {
	Name mstr `json:"name" toml:"name" yaml:"name"`
}

type muCfg struct {
	Items []muItem `json:"items" toml:"items" yaml:"items"`
	Nums  []int    `json:"nums"  toml:"nums"  yaml:"nums"`
}

func TestMetaArrayUniformityJSON(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameJSON: SourceBytes("{\"nums\":[1,2],\"items\":[{\"name\":\"a\"},{\"name\":\"b\"}]}\n")},
	})
	t.Chdir(string(root))

	_, _, meta, err := LoadWithMeta[muCfg](Spec{Name: configName("app"), Format: FormatJSON},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	assertArrayContainerAndElements(t, meta, "nums", 2, false)
	assertArrayOfObjectsShape(t, meta)
}

func TestMetaArrayUniformityYAML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("nums:\n  - 1\n  - 2\nitems:\n  - name: a\n  - name: b\n")},
	})
	t.Chdir(string(root))

	_, _, meta, err := LoadWithMeta[muCfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	assertArrayContainerAndElements(t, meta, "nums", 2, false)
	assertArrayOfObjectsShape(t, meta)
}
