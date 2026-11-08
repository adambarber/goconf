package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adambarber/gorefuse"

	"github.com/adambarber/goconf"
)

type probeConfig struct {
	Nested probeInner
	Region region
	Port   probePort
}

type probeInner struct {
	Host probeHost `json:"host" toml:"host" yaml:"host"`
}

type region string

type probeHost string

type probePort int

// probeCase is one refusal shape to render and judge.
type probeCase struct {
	name   probeName
	ext    probeExt
	body   probeBody
	format goconf.Format
}

type probeName string

type probeExt string

type probeBody string

type dirPath string

// probeFilePerm is the probe file's mode; gosec wants 0600 or less.
const probeFilePerm = 0o600

// probe extension constants: goconst wants literals deduplicated.
const (
	extJSON = probeExt(".json")
	extYAML = probeExt(".yaml")
	extTOML = probeExt(".toml")
)

// probeCases covers every user-facing refusal shape: syntax/parse, the
// four reportable vocabulary cases, and the structural refusals.
func probeCases() []probeCase {
	return []probeCase{
		{name: "json syntax", ext: extJSON, body: "{\n  \"region\": \"eu\"\n", format: goconf.FormatJSON},
		{name: "json empty", ext: extJSON, body: "", format: goconf.FormatJSON},
		{name: "yaml scanner", ext: extYAML, body: "region: eu\nretries: @x`broken\n", format: goconf.FormatYAML},
		{name: "toml parse err", ext: extTOML, body: "retries = zzq-secret-marker-8842\n", format: goconf.FormatTOML},
		{name: "yaml unknown field", ext: extYAML, body: "region: eu\nextra: 1\n", format: goconf.FormatYAML},
		{name: "yaml type mismatch", ext: extYAML, body: "port: not-a-number\n", format: goconf.FormatYAML},
		{name: "toml unknown field", ext: extTOML, body: "region = \"eu\"\nextra = 1\n", format: goconf.FormatTOML},
		{name: "json unknown field", ext: extJSON, body: `{"region":"eu","extra":true}`, format: goconf.FormatJSON},
		{name: "json type mismatch", ext: extJSON, body: `{"port": "not-a-number"}`, format: goconf.FormatJSON},
		{name: "toml type mismatch", ext: extTOML, body: "port = \"abc\"\n", format: goconf.FormatTOML},
		{name: "json not-a-mapping", ext: extJSON, body: "[1,2,3]", format: goconf.FormatJSON},
		{name: "yaml not-a-mapping", ext: extYAML, body: "- a\n- b\n", format: goconf.FormatYAML},
		{name: "yaml multi-doc", ext: extYAML, body: "region: eu\n---\nregion: us\n", format: goconf.FormatYAML},
		{name: "json duplicate key", ext: extJSON, body: "{\n  \"region\": \"eu\",\n  \"region\": \"us\"\n}\n", format: goconf.FormatJSON},
		{name: "yaml value echo", ext: extYAML, body: "port: \"secret-value-42\"\n", format: goconf.FormatYAML},
		{name: "json nested unknown", ext: extJSON, body: `{"nested":{"extra":1}}`, format: goconf.FormatJSON},
	}
}

func main() {
	dir, err := os.MkdirTemp("", "goconfprobe")
	if err != nil {
		fmt.Println("temp dir:", err)
		return
	}

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	root := dirPath(dir)

	for _, tc := range probeCases() {
		run(root, tc)
	}
}

// run writes one case's file, loads, and renders the refusal.
func run(dir dirPath, tc probeCase) {
	path := filepath.Join(string(dir), "hs"+string(tc.ext))

	if err := os.WriteFile(path, []byte(tc.body), probeFilePerm); err != nil {
		fmt.Println("write:", err)
		return
	}

	_, _, err := goconf.Load[probeConfig](goconf.Spec{Name: "hs", Format: tc.format},
		goconf.WithSearchPaths(goconf.DirPath(string(dir))))

	show(tc.name, err)

	_ = os.Remove(path)
}

// show renders one refusal for inspection.
func show(name probeName, err error) {
	fmt.Printf("=== %s\n", name)

	if err == nil {
		fmt.Println("  (no error)")
		return
	}

	fmt.Printf("  Error(): %s\n", err.Error())

	b, ok := errors.AsType[gorefuse.Because[goconf.Reason]](err)
	if ok {
		fmt.Printf("  Reason: %q Line:%d\n", b.Reason, b.What.Line)
		return
	}

	fmt.Println("  NOT a Because[Reason]!")
}
