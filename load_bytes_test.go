package goconf

import (
	"errors"
	"reflect"
	"testing"
)

type metaText string

type metaShape struct {
	Path        OccurrencePath
	Schema      SchemaPath
	Key         OccurrenceKey
	Line        int
	SourceIndex int
	Effective   bool
	ValueKind   reflect.Kind
	ValueText   metaText
}

func shapeOf(meta Meta) []metaShape {
	entries := meta.Entries()

	out := make([]metaShape, 0, len(entries))
	for _, e := range entries {
		out = append(out, metaShape{
			Path:        e.Path,
			Schema:      e.SchemaPath,
			Key:         e.Key,
			Line:        e.Line,
			SourceIndex: e.SourceIndex,
			Effective:   e.Effective,
			ValueKind:   e.Value.Kind(),
			ValueText:   metaText(e.Value.String()),
		})
	}

	return out
}

func TestLoadBytesDecodesWithoutFilesystem(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	data := SourceBytes(`{"region":"eu","nested":{"host":"h","port":1}}`)

	got, prov, err := LoadBytes[loadCfg](data, Spec{Name: configName("memory.json"), Format: FormatJSON})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	cfgEquals(t, got, loadCfgWant(0, "eu", nestedCfgWant("h", 1)))

	if len(prov) != 1 {
		t.Fatalf("provenance = %d entries, want 1", len(prov))
	}

	if prov[0].Path != FilePath("memory.json") || prov[0].Digest != digestOf(data) {
		t.Fatalf("provenance = %+v, want name and digest of the bytes", prov[0])
	}
}

func TestLoadBytesWithMetaMatchesDiskShape(t *testing.T) {
	data := SourceBytes("{\n  \"region\": \"eu\",\n  \"retries\": 3\n}\n")
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		dirLocal: {fileNameJSON: data},
	})
	chdirLocal(t, root)

	_, _, diskMeta, err := LoadWithMeta[loadCfg](Spec{Name: configName("app"), Format: FormatJSON})
	if err != nil {
		t.Fatalf("disk load: %v", err)
	}

	got, prov, bytesMeta, err := LoadBytesWithMeta[loadCfg](data, Spec{Name: configName("app"), Format: FormatJSON})
	if err != nil {
		t.Fatalf("bytes load: %v", err)
	}

	cfgEquals(t, got, loadCfgWant(3, "eu", nestedCfgWant("", 0)))

	if len(prov) != 1 || prov[0].Digest != digestOf(data) {
		t.Fatalf("provenance = %+v, want one digest of the bytes", prov)
	}

	diskShape, bytesShape := shapeOf(diskMeta), shapeOf(bytesMeta)
	if !reflect.DeepEqual(diskShape, bytesShape) {
		t.Fatalf("meta shape differs: disk %+v, bytes %+v", diskShape, bytesShape)
	}
}

func TestLoadBytesRefusesUnknownKeyWithoutProvenance(t *testing.T) {
	got, prov, meta, err := LoadBytesWithMeta[loadCfg](
		SourceBytes(`{"region":"eu","extra":1}`),
		Spec{Name: configName("memory.json"), Format: FormatJSON})

	if !errors.Is(err, ErrUnknownField) {
		t.Fatalf("err = %v, want ErrUnknownField", err)
	}

	if got.Region != "" || prov != nil || meta.Len() != 0 {
		t.Fatalf("refusal returned config=%+v prov=%v meta=%d", got, prov, meta.Len())
	}
}

func TestLoadBytesAppliesDefaults(t *testing.T) {
	got, _, err := LoadBytes[loadCfg](SourceBytes(`{}`),
		Spec{Name: configName("memory.json"), Format: FormatJSON},
		WithDefaults(loadCfgWant(7, "eu", nestedCfgWant("h", 1))))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	cfgEquals(t, got, loadCfgWant(7, "eu", nestedCfgWant("h", 1)))
}

func TestLoadBytesRequiresFormat(t *testing.T) {
	_, _, err := LoadBytes[loadCfg](SourceBytes(`{}`), Spec{Name: configName("memory.json"), Format: undetermined})

	if !errors.Is(err, ErrUnknownFormat) {
		t.Fatalf("err = %v, want ErrUnknownFormat", err)
	}
}
