package goconf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/adambarber/gorefuse"
)

type (
	pathwayRegion string
	pathwayPort   int
	pathwayConfig struct {
		Region pathwayRegion `json:"region" toml:"region" yaml:"region"`
		Port   pathwayPort   `json:"port"   toml:"port"   yaml:"port"`
	}
)

const (
	pathwayAppName  Name        = "app"
	pathwayPortFull pathwayPort = 8080
	pathwayByteCap              = 8
)

func TestPublicPathwayStrictLoad(t *testing.T) {
	_, _, err := LoadBytes[pathwayConfig](
		[]byte(`{"region":"eu","typo":1}`),
		Spec{Name: pathwayAppName, Format: FormatJSON},
	)

	if !errors.Is(err, ErrUnknownField) {
		t.Fatalf("err = %v, want ErrUnknownField", err)
	}
}

func TestPublicPathwayDefaultsAndNull(t *testing.T) {
	defaults := pathwayConfig{Region: "", Port: pathwayPortFull}

	missing, _, err := LoadBytes[pathwayConfig](
		[]byte(`{"region":"eu"}`),
		Spec{Name: pathwayAppName, Format: FormatJSON},
		WithDefaults(defaults),
	)
	if err != nil {
		t.Fatalf("missing load: %v", err)
	}

	explicit, _, err := LoadBytes[pathwayConfig](
		[]byte(`{"region":"eu","port":null}`),
		Spec{Name: pathwayAppName, Format: FormatJSON},
		WithDefaults(defaults),
	)
	if err != nil {
		t.Fatalf("explicit load: %v", err)
	}

	if missing.Port != pathwayPortFull || explicit.Port != 0 {
		t.Fatalf("port = %d/%d, want %d/0", missing.Port, explicit.Port, pathwayPortFull)
	}
}

func TestPublicPathwayLayering(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{
		fileNameJSON: SourceBytes(`{"region":"local"}`),
	})

	global := DirPath(filepath.Join(string(root), string(dirGlobal)))
	if err := os.Mkdir(string(global), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	writeFileAt(t, global, fileNameJSON, SourceBytes(`{"region":"global","port":8080}`))

	cfg, _, err := Load[pathwayConfig](
		Spec{Name: pathwayAppName, Format: FormatJSON},
		WithSearchPaths(append([]DirPath{global}, legPaths(root, dirLocal)...)...),
	)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Region != "local" || cfg.Port != pathwayPortFull {
		t.Fatalf("cfg = %+v, want local/%d", cfg, pathwayPortFull)
	}
}

func TestPublicPathwayProvenanceAndMeta(t *testing.T) {
	_, prov, meta, err := LoadBytesWithMeta[pathwayConfig](
		[]byte("{\n  \"region\": \"eu\",\n  \"port\": 8080\n}"),
		Spec{Name: pathwayAppName, Format: FormatJSON},
	)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(prov) != 1 || prov[0].Digest == "" {
		t.Fatalf("provenance = %+v, want one digest", prov)
	}

	occurrences := meta.ByPath(OccurrencePath("port"))
	if len(occurrences) != 1 || occurrences[0].Line != 3 || !occurrences[0].Effective {
		t.Fatalf("occurrences = %+v, want one effective line-3 record", occurrences)
	}
}

func TestPublicPathwayRefusalIdentity(t *testing.T) {
	_, _, err := LoadBytes[pathwayConfig](
		[]byte(`{"region":"eu","typo":1}`),
		Spec{Name: pathwayAppName, Format: FormatJSON},
	)

	because, ok := errors.AsType[gorefuse.Because[Reason]](err)
	if !ok || because.Reason != ErrUnknownField || because.What.File != gorefuse.File(pathwayAppName) {
		t.Fatalf("because = %+v, ok = %v", because, ok)
	}
}
