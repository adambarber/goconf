package goconf

import (
	"testing"
)

type mrsrv struct {
	Name  mstr     `json:"name"  toml:"name"  yaml:"name"`
	Ports []mrport `json:"ports" toml:"ports" yaml:"ports"`
}

const (
	mrSchemaPorts    SchemaPath = "servers.ports"
	mrSchemaPortsNum SchemaPath = "servers.ports.num"
)

type mrport struct {
	Num int `json:"num" toml:"num" yaml:"num"`
}

type mrcfg struct {
	Servers []mrsrv `json:"servers" toml:"servers" yaml:"servers"`
}

func TestLoadWithMetaTOMLNestedAOTIndexed(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("" +
			"[[servers]]\nname = \"s1\"\n" +
			"[[servers.ports]]\nnum = 80\n" +
			"[[servers.ports]]\nnum = 81\n" +
			"[[servers]]\nname = \"s2\"\n" +
			"[[servers.ports]]\nnum = 90\n")},
	})
	t.Chdir(string(root))

	got, _, meta, err := LoadWithMeta[mrcfg](Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if len(got.Servers) != 2 || got.Servers[0].Name != "s1" || len(got.Servers[0].Ports) != 2 || got.Servers[0].Ports[0].Num != 80 || got.Servers[0].Ports[1].Num != 81 || got.Servers[1].Name != "s2" || len(got.Servers[1].Ports) != 1 || got.Servers[1].Ports[0].Num != 90 {
		t.Fatalf("decoded=%#v", got)
	}

	assertNestedAOTPaths(t, meta)
}

func assertNestedAOTPaths(t *testing.T, meta Meta) {
	t.Helper()

	cases := []struct {
		path   OccurrencePath
		schema SchemaPath
		eff    bool
	}{
		{path: "servers[0]", schema: "servers", eff: true},
		{path: "servers[0].name", schema: "servers.name", eff: true},
		{path: "servers[0].ports[0]", schema: mrSchemaPorts, eff: true},
		{path: "servers[0].ports[0].num", schema: mrSchemaPortsNum, eff: true},
		{path: "servers[0].ports[1]", schema: mrSchemaPorts, eff: true},
		{path: "servers[0].ports[1].num", schema: mrSchemaPortsNum, eff: true},
		{path: "servers[1]", schema: "servers", eff: true},
		{path: "servers[1].name", schema: "servers.name", eff: true},
		{path: "servers[1].ports[0]", schema: mrSchemaPorts, eff: true},
		{path: "servers[1].ports[0].num", schema: mrSchemaPortsNum, eff: true},
	}

	for _, tc := range cases {
		es := meta.ByPath(tc.path)
		if len(es) == 0 {
			t.Fatalf("missing path %q in %#v", tc.path, meta.Entries())
		}

		if es[0].SchemaPath != tc.schema || es[0].Effective != tc.eff {
			t.Fatalf("path %q got schema=%q eff=%v", tc.path, es[0].SchemaPath, es[0].Effective)
		}
	}

	// no collisions at schema path level for specific indexed elements
	nums := meta.BySchemaPath(mrSchemaPortsNum)
	if len(nums) != 3 {
		t.Fatalf("%s entries=%d %#v", mrSchemaPortsNum, len(nums), nums)
	}
}

func TestLoadWithMetaTOMLInlineArrayElements(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameTOML: SourceBytes("nums = [10, 20, 30]\n")},
	})
	t.Chdir(string(root))

	type cfg struct {
		Nums []int `toml:"nums"`
	}

	got, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatTOML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if len(got.Nums) != 3 || got.Nums[0] != 10 || got.Nums[1] != 20 || got.Nums[2] != 30 {
		t.Fatalf("nums=%#v", got.Nums)
	}

	for i := range 3 {
		path := OccurrencePath("nums[" + string(rune('0'+i)) + "]")

		es := meta.ByPath(path)
		if len(es) != 1 {
			t.Fatalf("path %q entries=%#v all=%#v", path, es, meta.Entries())
		}

		if es[0].SchemaPath != "nums" || !es[0].Effective {
			t.Fatalf("path %q entry=%#v", path, es[0])
		}
	}
}

func TestLoadWithMetaYAMLMergeKeyEffective(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirGlobal: {fileNameYAML: SourceBytes("base: &base\n  x: 1\n  y: 2\na:\n  <<: *base\n  y: 3\n")},
	})
	t.Chdir(string(root))

	type inner struct {
		X int `yaml:"x"`
		Y int `yaml:"y"`
	}

	type cfg struct {
		Base inner `yaml:"base"`
		A    inner `yaml:"a"`
	}

	got, _, meta, err := LoadWithMeta[cfg](Spec{Name: configName("app"), Format: FormatYAML},
		loadOpts(globalPathOf(root))...)
	if err != nil {
		t.Fatalf("LoadWithMeta: %v", err)
	}

	if got.A.X != 1 || got.A.Y != 3 {
		t.Fatalf("a=%#v base=%#v", got.A, got.Base)
	}

	ax := meta.ByPath("a.x")
	if len(ax) == 0 {
		t.Fatalf("missing a.x, entries=%#v", meta.Entries())
	}

	if !ax[0].Effective {
		t.Fatalf("a.x not effective: %#v", ax[0])
	}

	ay := meta.ByPath("a.y")
	if len(ay) == 0 || !ay[0].Effective {
		t.Fatalf("a.y entries=%#v", ay)
	}
}
