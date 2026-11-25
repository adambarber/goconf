package goconf

import "testing"

func TestLoadWithMetaYAMLArrayOccurrences(t *testing.T) {
	_, meta := loadMetaGlobal[metaItemsCfg](t, SourceBytes("items:\n  - name: a\n  - name: b\n"), FormatYAML)

	sn := meta.BySchemaPath("items.name")
	if len(sn) != 2 || sn[0].Path != metaPathItems0Name || sn[1].Path != metaPathItems1Name {
		t.Fatalf("entries=%#v", meta.Entries())
	}

	if sn[0].Line != 2 || sn[1].Line != 3 {
		t.Fatalf("lines=%#v", sn)
	}
}

func TestLoadWithMetaExtrasTOML(t *testing.T) {
	_, meta := loadMetaGlobal[metaExtrasTOMLCfg](t, SourceBytes("name = \"n\"\nextra = \"x\"\n"), FormatTOML)

	if len(meta.BySchemaPath("extra")) != 1 {
		t.Fatalf("entries=%#v", meta.Entries())
	}
}

func TestLoadWithMetaDynamicYAML(t *testing.T) {
	_, meta := loadMetaGlobal[metaDynamicJSONCfg](t, SourceBytes("env:\n  A: \"1\"\n"), FormatYAML)

	if len(meta.ByPath("env.A")) != 1 {
		t.Fatalf("entries=%#v", meta.Entries())
	}
}

func TestLoadWithMetaDefaultsEffective(t *testing.T) {
	got, meta := loadMetaGlobal[metaRegionRetriesCfg](t, SourceBytes("{\"region\":\"eu\"}\n"), FormatJSON,
		WithDefaults(metaRegionRetriesCfg{Region: "d", Retries: 5}))

	if got.Region != "eu" || got.Retries != 5 {
		t.Fatalf("got=%+v", got)
	}

	rs := meta.BySchemaPath("region")
	if len(rs) != 1 || !rs[0].Effective || !rs[0].Value.IsValid() || rs[0].Value.String() != "eu" {
		t.Fatalf("region=%#v", rs)
	}
}
